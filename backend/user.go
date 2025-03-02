package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/unrolled/render"
	"golang.org/x/crypto/argon2"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type UserService struct {
	db          *gorm.DB
	emailDialer *gomail.Dialer
	privateKey  crypto.PrivateKey
	publicKey   crypto.PublicKey
}

type User struct {
	UUID               uuid.UUID `gorm:"primaryKey;unique;type:uuid"`
	FirstName          string    `gorm:"not null"`
	LastName           string    `gorm:"not null"`
	Password           []byte    `gorm:"not null"`
	Email              string    `gorm:"unique,not null"`
	RegistrationToken  string
	RefreshToken       string
	AccessToken        string
	Amount             float64
	RefreshTokenExpiry sql.NullTime
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}

type UnverifiedUser struct {
	UUID              uuid.UUID `gorm:"primaryKey;unique;type:uuid"`
	FirstName         string    `gorm:"not null"`
	LastName          string    `gorm:"not null"`
	Email             string    `gorm:"not null"`
	VerificationToken string    `gorm:"not null"`
	Password          []byte    `gorm:"not null"`
	Expiry            time.Time `gorm:"not null"`
	Activated         sql.NullTime
}

type NewUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type Response struct {
	Successful bool              `json:"success"`
	Error      map[string]string `json:"error,omitempty"`
}

type LoginRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	RegistrationToken string `json:"registration_token"` // FCM
}

func NewUserService(db *gorm.DB, emailDialer *gomail.Dialer, privateKeyPath, publicKeyPath string) *UserService {
	// openssl ecparam -name prime256v1 -genkey -noout -out private_key.pem
	pem, err := os.ReadFile(privateKeyPath)
	if err != nil {
		panic(err)
	}

	privateKey, err := jwt.ParseECPrivateKeyFromPEM(pem)
	if err != nil {
		panic(err)
	}

	// openssl ec -in private_key.pem -pubout > public_key.pem
	pem, err = os.ReadFile(publicKeyPath)
	if err != nil {
		panic(err)
	}

	publicKey, err := jwt.ParseECPublicKeyFromPEM(pem)
	if err != nil {
		panic(err)
	}

	if err = db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}

	if err = db.AutoMigrate(&UnverifiedUser{}); err != nil {
		panic(err)
	}

	service := &UserService{db, emailDialer, privateKey, publicKey}

	http.HandleFunc("/register", service.HandleRegister)
	http.HandleFunc("/verify", service.VerifyUser)
	http.HandleFunc("/login", service.Login)
	http.HandleFunc("/userdata", service.GetUserData)

	return service
}

// RenderErrorTemplate Utility function to render error templates
func RenderErrorTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	if err := render.New().HTML(w, http.StatusOK, templateName, data); err != nil {
		log.Fatal(err)
	}
}

// GenerateVerificationToken Utility function to generate random tokens
func GenerateVerificationToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

// HashPassword Utility function to hash passwords using argon2
func HashPassword(password string, salt string) []byte {
	return argon2.IDKey([]byte(password), []byte(salt), 1, 64*1024, 4, 32)
}

// EmailExists Check if a user with the given email exists in the database
func (service *UserService) EmailExists(email string) bool {
	return service.db.First(&User{}, "email = ?", email).RowsAffected > 0
}

// UnverifiedUserExists Check if an unverified user with the given email and token exists in the database
func (service *UserService) UnverifiedUserExists(email, verificationToken string) bool {
	var unverifiedUser UnverifiedUser
	return service.db.First(&unverifiedUser, "email = ? AND verification_token = ?", email, verificationToken).Error == nil
}

// SendVerificationEmail Create and broadcast a verification email
func (service *UserService) SendVerificationEmail(user UnverifiedUser, host string) {
	tpl, err := ParseTemplate("verify-email.tmpl", fmt.Sprintf("https://%s/verify?token=%s&email=%s", host, user.VerificationToken, url.QueryEscape(user.Email)))
	if err != nil {
		log.Fatalln(err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", service.emailDialer.Username)
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", "Verify Email")
	m.SetBody("text/html", tpl)

	if err = service.emailDialer.DialAndSend(m); err != nil {
		log.Println(err)
	}
}

// ParseTemplate Parse and execute a template file
func ParseTemplate(templateName, data string) (string, error) {
	t, err := template.New(templateName).ParseFiles("./templates/" + templateName)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func ShowError(w http.ResponseWriter) {
	RenderErrorTemplate(w, "error", nil)
}

func (service *UserService) VerifyUser(w http.ResponseWriter, r *http.Request) {
	SetCors(&w)

	if r.Method != http.MethodGet {
		return
	}

	email := strings.ToLower(r.URL.Query().Get("email"))
	token := r.URL.Query().Get("token")

	if !service.UnverifiedUserExists(email, token) {
		RenderErrorTemplate(w, "verification-failed", map[string]string{"Error": "we couldn't find your verification link in our database."})
		return
	}

	var unverifiedUser UnverifiedUser
	service.db.First(&unverifiedUser, "email = ? AND verification_token = ?", email, token)

	if unverifiedUser.Activated.Valid && unverifiedUser.Activated.Time.Before(time.Now()) {
		RenderErrorTemplate(w, "successfully-verified", map[string]string{"Name": unverifiedUser.FirstName})
		return
	}

	// Check if verification token has expired
	if time.Now().After(unverifiedUser.Expiry) {
		RenderErrorTemplate(w, "verification-failed", map[string]string{
			"Name":  unverifiedUser.FirstName,
			"Error": "too much time has passed since the verification email was sent.",
		})
		return
	}

	// Check if the user already exists
	if service.EmailExists(email) {
		RenderErrorTemplate(w, "verification-failed", map[string]string{
			"Name":  unverifiedUser.FirstName,
			"Error": "an account with this email already exists.",
		})
		return
	}

	// Create the new user
	user := &User{
		FirstName: unverifiedUser.FirstName,
		LastName:  unverifiedUser.LastName,
		UUID:      unverifiedUser.UUID,
		Password:  unverifiedUser.Password,
		Email:     unverifiedUser.Email,
		Amount:    10,
	}

	if err := service.db.Create(user).Error; err != nil {
		log.Println(err)
		ShowError(w)
		return
	}

	unverifiedUser.Activated = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	// Mark the link as used
	if err := service.db.Save(&unverifiedUser).Error; err != nil {
		log.Println(err)
		ShowError(w)
		return
	}

	// Send the success response
	err := render.New().HTML(w, http.StatusOK, "successfully-verified", map[string]string{
		"Name": unverifiedUser.FirstName,
	})
	if err != nil {
		log.Println(err)
	}
}

func (service *UserService) HandleRegister(w http.ResponseWriter, r *http.Request) {
	SetCors(&w)

	if r.Method != http.MethodPost {
		return
	}

	var request NewUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Println(err)
		InvalidJSONError(w, err)
		return
	}

	user, err := service.ProcessUser(&request, w, r.Host)
	if err != nil {
		log.Println(err)
		return
	}

	// Create the unverified user
	if err := service.db.Create(user).Error; err != nil {
		response := NewError(31, "Error adding user to database", err.Error())
		RenderJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	RenderJSONResponse(w, http.StatusCreated, &Response{Successful: true})
}

func (service *UserService) ProcessUser(request *NewUserRequest, w http.ResponseWriter, host string) (*UnverifiedUser, error) {
	email := strings.ToLower(request.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		response := NewError(32, "Invalid email", err.Error())
		RenderJSONResponse(w, http.StatusBadRequest, response)
		return nil, fmt.Errorf("invalid email: %s", email)
	}

	// Check if email already exists
	if service.EmailExists(email) {
		response := NewError(33, "Account already exists with email", "Account already exists with email: "+email)
		RenderJSONResponse(w, http.StatusConflict, response)
		return nil, fmt.Errorf("account already exists with email: %s", email)
	}

	userID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	// Generate the unverified user
	user := &UnverifiedUser{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		UUID:      userID,
		Email:     email,
		Password:  HashPassword(request.Password, userID.String()),
	}

	// Generate token and expiry
	token, err := GenerateVerificationToken()
	if err != nil {
		return nil, err
	}

	user.VerificationToken = token
	user.Expiry = time.Now().Add(time.Hour * 24 * 7)

	// Send verification email asynchronously
	go service.SendVerificationEmail(*user, host)

	return user, nil
}

type LoginResponse struct {
	Successful  bool   `json:"success"`
	AccessToken string `json:"access_token"`
	UUID        string `json:"uuid"`
}

func (service *UserService) Login(w http.ResponseWriter, r *http.Request) {
	SetCors(&w)

	if r.Method != http.MethodPost {
		return
	}

	var request LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Println(err)
		InvalidJSONError(w, err)
		return
	}

	user := &User{
		Email: request.Email,
	}

	if !service.EmailExists(request.Email) {
		response := NewError(13, "Email not found.", "Email not found: "+request.Email)
		RenderJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	service.db.First(user, "email = ?", request.Email)

	if bytes.Equal(HashPassword(request.Password, user.UUID.String()), user.Password) {
		accessToken, err := service.GenerateAccessToken(user)
		if err != nil {
			log.Println(err)
			response := NewError(14, "Error signing token", err.Error())
			RenderJSONResponse(w, http.StatusInternalServerError, response)
			return
		}

		refreshToken, err := uuid.NewRandom()
		if err != nil {
			log.Println(err)
			response := NewError(15, "Error creating refresh token", err.Error())
			RenderJSONResponse(w, http.StatusInternalServerError, response)
			return
		}

		user.RegistrationToken = request.RegistrationToken
		user.RefreshToken = refreshToken.String()
		user.AccessToken = accessToken
		user.RefreshTokenExpiry = sql.NullTime{
			Time:  time.Now().Add(time.Hour * 24 * 7),
			Valid: true,
		}

		service.db.Save(user)

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    user.RefreshToken,
			HttpOnly: true,
			Domain:   r.Host,
		})

		RenderJSONResponse(w, 200, &LoginResponse{true, accessToken, user.UUID.String()})
		return
	}

	response := NewError(15, "Invalid password", "Invalid password")
	RenderJSONResponse(w, http.StatusBadRequest, response)
	return
}

func (service *UserService) GenerateAccessToken(user *User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256,
		jwt.MapClaims{
			"uuid":   user.UUID,
			"expiry": time.Now().Add(time.Hour).Unix(),
		})

	signedString, err := token.SignedString(service.privateKey)
	if err != nil {
		return "", err
	}

	return signedString, nil
}

type RefreshTokenResponse struct {
	AccessToken string    `json:"access_token"`
	Expiry      time.Time `json:"expiry"`
}

func (service *UserService) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}

	email := strings.ToLower(r.URL.Query().Get("email"))

	if !service.EmailExists(email) {
		response := NewError(13, "Email not found.", "Email not found: "+email)
		RenderJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		response := NewError(14, "Invalid refresh token.", err.Error())
		RenderJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	user := &User{
		Email:        email,
		RefreshToken: refreshToken.Value,
	}

	if service.db.First(&user, "email = ?", email).RowsAffected == 0 {
		response := NewError(14, "Invalid refresh token.", "Invalid refresh token with email: "+email)
		RenderJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	if time.Now().After(user.RefreshTokenExpiry.Time) {
		response := NewError(15, "Refresh token is expired.", "Refresh token is expired. Please login again.")
		RenderJSONResponse(w, http.StatusUnauthorized, response)
		return
	}

	accessToken, err := service.GenerateAccessToken(user)
	if err != nil {
		log.Println(err)
		response := NewError(14, "Error signing token", err.Error())
		RenderJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	RenderJSONResponse(w, 200, &RefreshTokenResponse{accessToken, time.Now().Add(time.Minute * 15)})
}

func (service *UserService) AuthenticateRequest(accessToken string) (*User, *Response) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return service.publicKey, nil
	})

	if err != nil {
		return nil, NewError(14, "Error parsing token", err.Error())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		expiryUnix, ok := claims["expiry"].(float64)
		if !ok {
			return nil, NewError(15, "Error parsing token expiry", "Error parsing token expiry: ")
		}

		expiry := time.Unix(int64(expiryUnix), 0)
		if time.Now().After(expiry) {
			return nil, NewError(15, "Access token is expired", "Access token is expired")
		}

		userID, ok := claims["uuid"].(string)
		if !ok {
			return nil, NewError(16, "Error parsing token uuid", "Error parsing token uuid")
		}

		user := &User{}
		if service.db.First(&user, "uuid = ?", userID).Error != nil {
			return nil, NewError(17, "User not found.", "User not found with provided email or has incorrect access token.")
		}

		return user, nil
	}

	return nil, NewError(15, "Invalid token", "Invalid token")
}

type UserDataRequest struct {
	AccessToken string `json:"access_token"`
}

type UserDataResponse struct {
	Success   bool    `json:"success"`
	Balance   float64 `json:"balance"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     string  `json:"email"`
}

func (service *UserService) GetUserData(w http.ResponseWriter, r *http.Request) {
	SetCors(&w)

	if r.Method != http.MethodPost {
		return
	}

	request := &UserDataRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		InvalidJSONError(w, err)
	}

	user, authResponse := service.AuthenticateRequest(request.AccessToken)
	if authResponse != nil {
		RenderJSONResponse(w, http.StatusInternalServerError, authResponse)
		return
	}

	RenderJSONResponse(w, 200, &UserDataResponse{true, user.Amount, user.FirstName, user.LastName, user.Email})
}

func (service *UserService) GetUser(uuid string) (*User, error) {
	user := &User{}
	return user, service.db.First(&user, "uuid = ?", uuid).Error
}
