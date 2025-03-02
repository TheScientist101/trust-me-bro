package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"math/rand"
	"net/http"
	"slices"
	"time"

	"gorm.io/gorm"
)

type PendingGame struct {
	ID                 uuid.UUID      `gorm:"primarykey" json:"id"`
	CreatedAt          time.Time      `json:"-"`
	UpdatedAt          time.Time      `json:"-"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
	PlayerOne          uuid.UUID      `gorm:"type:uuid" json:"-"`
	PlayerOneMove      string         `json:"-"`
	PlayerTwo          uuid.UUID      `gorm:"type:uuid" json:"-"`
	PlayerTwoMove      string         `json:"-"`
	Completed          bool           `gorm:"type:bool" json:"-"`
	Result             bool           `gorm:"type:text" json:"-"`
	TransactionID      uuid.UUID      `gorm:"type:uuid" json:"-"`
	PlayerOneIntention bool           `json:"-"`
}

type PendingGamesRequest struct {
	AccessToken string `json:"access_token"`
}

type PendingGamesResponse struct {
	Success bool     `json:"success"`
	GameIDs []string `json:"games"`
}

func (ts *TransactionService) GetPendingGames(w http.ResponseWriter, r *http.Request) {
	SetCors(&w)

	if r.Method != http.MethodPost {
		return
	}

	request := &PendingGamesRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		InvalidJSONError(w, err)
		return
	}

	user, authResponse := ts.us.AuthenticateRequest(request.AccessToken)
	if authResponse != nil {
		RenderJSONResponse(w, http.StatusInternalServerError, authResponse)
		return
	}

	var games []PendingGame
	var responses []TransactionResponse
	ts.db.Model(&TransactionResponse{}).Where("user_id = ?", user.UUID).Find(&responses)
	for _, response := range responses {
		var currGames []PendingGame
		ts.db.Where("player_one = ? OR player_two = ?", response.UserID, response.UserID).Find(&currGames)
		games = append(games, currGames...)
	}

	gameIDs := make([]string, 0, len(games))
	for _, game := range games {
		gameIDs = append(gameIDs, game.ID.String())
	}

	slices.Sort(gameIDs)
	slices.Compact(gameIDs)

	RenderJSONResponse(w, http.StatusOK, &PendingGamesResponse{
		Success: true,
		GameIDs: gameIDs,
	})
}

type PlayMoveRequest struct {
	GameID      uuid.UUID `json:"game_id"`
	AccessToken string    `json:"access_token"`
	Move        string    `json:"move"`
}

// SolveGame Returns whether player one would win
func SolveGame(moveOne, moveTwo string) bool {
	if moveOne == moveTwo {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		return rand.Intn(2) == 0
	}

	if (moveOne == "rock" && moveTwo == "scissors") ||
		(moveOne == "scissors" && moveTwo == "paper") ||
		(moveOne == "paper" && moveTwo == "rock") {
		return true // Player one wins
	}

	return false
}

func (ts *TransactionService) PlayMove(w http.ResponseWriter, r *http.Request) {
	SetCors(&w)

	if r.Method != http.MethodPost {
		return
	}

	request := &PlayMoveRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		InvalidJSONError(w, err)
		return
	}

	user, authResponse := ts.us.AuthenticateRequest(request.AccessToken)
	if authResponse != nil {
		RenderJSONResponse(w, http.StatusInternalServerError, authResponse)
		return
	}

	game := &PendingGame{}
	if ts.db.First(game, "id = ?", request.GameID).RowsAffected == 0 {
		RenderJSONResponse(w, http.StatusNotFound, NewError(12, "Game not found", "Game not found"))
		return
	}

	if !slices.Contains([]string{"rock", "paper", "scissors"}, request.Move) {
		RenderJSONResponse(w, http.StatusNotFound, NewError(12, "Invalid move", "Invalid move"))
	}

	if game.PlayerOne == user.UUID {
		if game.PlayerOneMove != "" {
			RenderJSONResponse(w, http.StatusOK, NewError(16, "Move already played", "Move already played"))
			return
		}

		game.PlayerOneMove = request.Move
	} else if game.PlayerTwo == user.UUID {
		if game.PlayerTwoMove != "" {
			RenderJSONResponse(w, http.StatusOK, NewError(16, "Move already played", "Move already played"))
			return
		}

		game.PlayerTwoMove = request.Move
	}

	if game.PlayerOneMove != "" && game.PlayerTwoMove != "" {
		game.Completed = true
		game.Result = SolveGame(game.PlayerOneMove, game.PlayerTwoMove)

		transaction := &Transaction{}
		ts.db.First(transaction, "uuid = ?", game.TransactionID)

		transaction.Approved = game.PlayerOneIntention && game.Result
		ts.db.Save(transaction)
	}

	ts.db.Save(game)
}
