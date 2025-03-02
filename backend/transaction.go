package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	UUID      uuid.UUID `gorm:"primaryKey;type:uuid"`
	From      User      `gorm:"ForeignKey:UUID"`
	To        User      `gorm:"ForeignKey:UUID"`
	Amount    float64   `gorm:"not null"`
	Approved  bool      `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type TransactionService struct {
	us     *UserService
	db     *gorm.DB
	client *messaging.Client
	ctx    context.Context
}

func NewTransactionService(db *gorm.DB, us *UserService, app *firebase.App) (*TransactionService, error) {
	ctx := context.Background()

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	ts := &TransactionService{us: us, db: db, client: client, ctx: ctx}

	if err = db.AutoMigrate(&Transaction{}); err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&TransactionResponse{}); err != nil {
		return nil, err
	}

	http.HandleFunc("/send", ts.Send)

	return ts, nil
}

type SendRequest struct {
	AccessToken string  `json:"access_token"`
	To          string  `json:"to"` // uuid
	Amount      float64 `json:"amount"`
}

func NewTransaction(from, to *User, amount float64) (*Transaction, error) {
	id, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	return &Transaction{
		UUID:     id,
		From:     *from,
		To:       *to,
		Amount:   amount,
		Approved: false,
	}, nil
}

func (ts *TransactionService) Send(w http.ResponseWriter, r *http.Request) {
	SetCors(&w)

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

	recipient, err := ts.us.GetUser(request.To)

	transaction, err := NewTransaction(sender, recipient, request.Amount)
	if err != nil {
		InvalidJSONError(w, err)
		return
	}

	ts.db.Save(transaction)

	message := &messaging.Message{
		Data: map[string]string{
			"transactionID": transaction.UUID.String(),
			"from":          sender.UUID.String(),
			"to":            recipient.UUID.String(),
		},
		Topic: "transaction",
	}

	_, err = ts.client.Send(ts.ctx, message)
	if err != nil {
		RenderJSONResponse(w, http.StatusInternalServerError, NewError(73, "Failed to send notifications.", err.Error()))
	}
}
