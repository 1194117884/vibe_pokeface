package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

type ScoreStore interface {
	GetUserBalance(ctx context.Context, userID int64) (int, error)
	GetScoreHistory(ctx context.Context, userID int64, limit int) ([]model.ScoreRecord, error)
	SaveScore(ctx context.Context, userID int64, gameType string, amount, balance int, reason string) error
}

type ScoreHandler struct {
	store ScoreStore
}

func NewScoreHandler(store ScoreStore) *ScoreHandler {
	return &ScoreHandler{store: store}
}

type balanceResponse struct {
	UserID  int64               `json:"user_id"`
	Balance int                 `json:"balance"`
	History []model.ScoreRecord `json:"history"`
}

type adjustRequest struct {
	UserID   int64  `json:"user_id"`
	Amount   int    `json:"amount"`
	Reason   string `json:"reason"`
	GameType string `json:"game_type"`
}

func (h *ScoreHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"user_id required"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	balance, err := h.store.GetUserBalance(ctx, userID)
	if err != nil {
		balance = 0
	}

	history, err := h.store.GetScoreHistory(ctx, userID, 20)
	if err != nil {
		history = []model.ScoreRecord{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balanceResponse{
		UserID:  userID,
		Balance: balance,
		History: history,
	})
}

func (h *ScoreHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	var req adjustRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.UserID == 0 || req.Amount == 0 {
		http.Error(w, `{"error":"user_id and amount required"}`, http.StatusBadRequest)
		return
	}
	if req.GameType == "" {
		req.GameType = "admin"
	}

	ctx := r.Context()
	currentBalance, err := h.store.GetUserBalance(ctx, req.UserID)
	if err != nil {
		currentBalance = 0
	}
	newBalance := currentBalance + req.Amount

	if err := h.store.SaveScore(ctx, req.UserID, req.GameType, req.Amount, newBalance, req.Reason); err != nil {
		http.Error(w, `{"error":"failed to adjust score"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"balance": newBalance,
	})
}
