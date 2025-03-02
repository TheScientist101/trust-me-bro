package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
	"net/http"
)

type TransactionResponse struct {
	UUID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID      uuid.UUID `gorm:"type:uuid"`
	Trust       bool      `gorm:"not null"`
	Transaction uuid.UUID `gorm:"type:uuid"`
}

type VoteRequest struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	AccessToken   string    `json:"access_token"`
	Trust         bool      `json:"trust"`
}

func (ts *TransactionService) ReceiveVote(w http.ResponseWriter, r *http.Request) {
	SetCors(&w)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		InvalidJSONError(w, err)
		return
	}

	user, authResponse := ts.us.AuthenticateRequest(request.AccessToken)
	if authResponse != nil {
		RenderJSONResponse(w, http.StatusInternalServerError, authResponse)
		return
	}

	id, err := uuid.NewRandom()
	if err != nil {
		RenderJSONResponse(w, http.StatusInternalServerError, NewError(16, "Error generating vote UUID", err.Error()))
		return
	}

	transaction := &Transaction{}
	if ts.db.First(transaction, "uuid = ?", request.TransactionID).RowsAffected == 0 {
		RenderJSONResponse(w, http.StatusNotFound, NewError(17, "Transaction not found", "Transaction not found"))
		return
	}

	vote := &TransactionResponse{
		UUID:        id,
		UserID:      user.UUID,
		Trust:       request.Trust,
		Transaction: request.TransactionID,
	}

	if transaction.FirstVote.String() == uuid.Nil.String() {
		transaction.FirstVote = id
	} else if transaction.SecondVote.String() == uuid.Nil.String() {
		transaction.SecondVote = id

		otherVote := &TransactionResponse{}
		ts.db.First(otherVote, "uuid = ?", transaction.FirstVote)

		if otherVote.Trust == vote.Trust {
			transaction.Approved = vote.Trust
		} else {
			game := &PendingGame{
				ID:                 uuid.New(),
				PlayerOne:          otherVote.UserID,
				PlayerTwo:          vote.UserID,
				TransactionID:      transaction.UUID,
				PlayerOneIntention: otherVote.Trust,
			}

			transaction.GameID = game.ID

			ts.db.Create(game)

			otherUser, err := ts.us.GetUser(otherVote.UserID.String())
			if err != nil {
				RenderJSONResponse(w, http.StatusInternalServerError, NewError(16, "Error getting user from database", err.Error()))
				return
			}

			users := []User{*user, *otherUser}
			m := gomail.NewMessage()
			m.SetHeader("From", ts.us.emailDialer.Username)
			addresses := make([]string, len(users))
			for i, user := range users {
				addresses[i] = m.FormatAddress(user.Email, user.FirstName)
			}
			m.SetHeader("To", addresses...)
			m.SetHeader("Subject", "If you win, democracy prevails...")
			m.SetBody("text/plain", "A new game has started, please login to your dashboard to play.")
			err = ts.us.emailDialer.DialAndSend(m)
			if err != nil {
				RenderJSONResponse(w, http.StatusInternalServerError, NewError(31, "Send failed", "Send failed"))
				return
			}
		}
	} else {
		RenderJSONResponse(w, http.StatusGone, NewError(10, "Votes already completed", "Votes already completed"))
		return
	}

	ts.db.Save(transaction)
	ts.db.Save(vote)
}
