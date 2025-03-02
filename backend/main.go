package main

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := gorm.Open(postgres.New(
		postgres.Config{
			DSN:                  os.Getenv("DATABASE_URL"),
			PreferSimpleProtocol: true,
		}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	emailPort, err := strconv.Atoi(os.Getenv("EMAIL_HOST_PORT"))
	if err != nil {
		panic(err)
	}

	d := gomail.NewDialer(os.Getenv("EMAIL_HOST"), emailPort, os.Getenv("EMAIL_ADDRESS"), os.Getenv("EMAIL_PASSWORD"))

	us := NewUserService(
		db,
		d,
		os.Getenv("PRIVATE_KEY_PATH"),
		os.Getenv("PUBLIC_KEY_PATH"),
	)

	opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDS_PATH"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic(err)
	}

	_, err = NewTransactionService(db, us, app)
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
