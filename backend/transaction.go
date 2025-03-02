package main

import (
	"encoding/json"
	"gopkg.in/gomail.v2"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	UUID       uuid.UUID `gorm:"primaryKey;type:uuid"`
	FromID     uuid.UUID `gorm:"type:uuid"`
	ToID       uuid.UUID `gorm:"type:uuid"`
	Amount     float64   `gorm:"not null"`
	FirstVote  uuid.UUID `gorm:"type:uuid"`
	SecondVote uuid.UUID `gorm:"type:uuid"`
	Approved   bool      `gorm:"not null"`
	GameID     uuid.UUID `gorm:"type:uuid"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type TransactionService struct {
	us *UserService
	db *gorm.DB
}

func NewTransactionService(db *gorm.DB, us *UserService) (*TransactionService, error) {
	ts := &TransactionService{us: us, db: db}

	if err := db.AutoMigrate(&Transaction{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&TransactionResponse{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&PendingGame{}); err != nil {
		return nil, err
	}

	http.HandleFunc("/send", ts.Send)
	http.HandleFunc("/pendingGames", ts.GetPendingGames)
	http.HandleFunc("/play", ts.PlayMove)
	http.HandleFunc("/vote", ts.ReceiveVote)

	return ts, nil
}

type SendRequest struct {
	AccessToken string  `json:"access_token"`
	To          string  `json:"to"` // email
	Amount      float64 `json:"amount"`
}

func NewTransaction(from, to uuid.UUID, amount float64) (*Transaction, error) {
	id, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	return &Transaction{
		UUID:     id,
		FromID:   from,
		ToID:     to,
		Amount:   amount,
		Approved: false,
	}, nil
}

func (ts *TransactionService) Send(w http.ResponseWriter, r *http.Request) {
	SetCors(&w)

	if r.Method != "POST" {
		return
	}

	request := &SendRequest{}

	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		InvalidJSONError(w, err)
		return
	}

	sender, authResponse := ts.us.AuthenticateRequest(request.AccessToken)
	if authResponse != nil {
		RenderJSONResponse(w, http.StatusUnauthorized, authResponse)
		return
	}

	if request.Amount > sender.Amount {
		RenderJSONResponse(w, http.StatusPaymentRequired, NewError(69, "You're broke", "Yeah, you don't have that much money"))
		return
	}

	recipient := &User{}
	if ts.db.First(recipient, "email = ?", request.To).RowsAffected == 0 {
		RenderJSONResponse(w, http.StatusNotFound, NewError(31, "Recipient not found", "Recipient not found"))
		return
	}

	if rand.Intn(100) < 10 {
		ts.db.Offset(rand.Intn(int(ts.db.Model(&User{}).RowsAffected))).First(recipient)
		ts.db.Offset(-1)
	}

	transaction, err := NewTransaction(sender.UUID, recipient.UUID, request.Amount)
	if err != nil {
		InvalidJSONError(w, err)
		return
	}

	sender.Amount -= request.Amount

	ts.db.Save(sender)
	ts.db.Save(transaction)

	var users []User

	ts.db.Model(&User{}).Find(&users)

	m := gomail.NewMessage()
	m.SetHeader("From", ts.us.emailDialer.Username)
	addresses := make([]string, len(users))
	for i, user := range users {
		addresses[i] = m.FormatAddress(user.Email, user.FirstName)
	}
	m.SetHeader("To", addresses...)
	m.SetHeader("Subject", "New Block to Add to the Chain")
	m.SetBody("text/plain", os.Getenv("CLIENT_HOST")+"voting?id="+transaction.UUID.String())
	err = ts.us.emailDialer.DialAndSend(m)
	if err != nil {
		RenderJSONResponse(w, http.StatusInternalServerError, NewError(31, "Send failed", "Send failed"))
		return
	}
}
