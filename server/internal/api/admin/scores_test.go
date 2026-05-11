package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

type mockScoreAdminStore struct {
	getUserBalanceFn  func(ctx context.Context, userID int64) (int, error)
	getScoreHistoryFn func(ctx context.Context, userID int64, limit int) ([]model.ScoreRecord, error)
	saveScoreFn       func(ctx context.Context, userID int64, gameType string, amount, balance int, reason string) error
}

func (m *mockScoreAdminStore) GetUserBalance(ctx context.Context, userID int64) (int, error) {
	return m.getUserBalanceFn(ctx, userID)
}
func (m *mockScoreAdminStore) GetScoreHistory(ctx context.Context, userID int64, limit int) ([]model.ScoreRecord, error) {
	return m.getScoreHistoryFn(ctx, userID, limit)
}
func (m *mockScoreAdminStore) SaveScore(ctx context.Context, userID int64, gameType string, amount, balance int, reason string) error {
	return m.saveScoreFn(ctx, userID, gameType, amount, balance, reason)
}

func TestScoreGetBalance(t *testing.T) {
	store := &mockScoreAdminStore{
		getUserBalanceFn: func(ctx context.Context, uid int64) (int, error) { return 1000, nil },
		getScoreHistoryFn: func(ctx context.Context, uid int64, limit int) ([]model.ScoreRecord, error) {
			return []model.ScoreRecord{
				{UserID: 1, GameType: "doudizhu", Amount: 100, Balance: 1000, Reason: "game win"},
			}, nil
		},
	}
	handler := NewScoreHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/scores?user_id=1", nil)
	w := httptest.NewRecorder()
	handler.GetBalance(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp balanceResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Balance != 1000 {
		t.Errorf("balance = %d, want 1000", resp.Balance)
	}
	if len(resp.History) != 1 {
		t.Errorf("history len = %d, want 1", len(resp.History))
	}
	if resp.UserID != 1 {
		t.Errorf("user_id = %d, want 1", resp.UserID)
	}
}

func TestScoreGetBalance_MissingUserID(t *testing.T) {
	handler := NewScoreHandler(&mockScoreAdminStore{})

	req := httptest.NewRequest("GET", "/api/admin/scores", nil)
	w := httptest.NewRecorder()
	handler.GetBalance(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestScoreAdjust_Success(t *testing.T) {
	store := &mockScoreAdminStore{
		getUserBalanceFn: func(ctx context.Context, uid int64) (int, error) { return 500, nil },
		saveScoreFn: func(ctx context.Context, uid int64, gt string, amount, balance int, reason string) error {
			if amount != 200 {
				t.Errorf("amount = %d, want 200", amount)
			}
			if balance != 700 {
				t.Errorf("balance = %d, want 700", balance)
			}
			return nil
		},
	}
	handler := NewScoreHandler(store)

	body := map[string]interface{}{
		"user_id": 1, "amount": 200, "reason": "admin bonus", "game_type": "admin",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/scores/adjust", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Adjust(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %v, want ok", resp["status"])
	}
	if resp["balance"] != float64(700) {
		t.Errorf("balance = %v, want 700", resp["balance"])
	}
}

func TestScoreAdjust_MissingFields(t *testing.T) {
	handler := NewScoreHandler(&mockScoreAdminStore{})

	body := map[string]interface{}{"user_id": 1}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/scores/adjust", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Adjust(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestScoreAdjust_NegativeAmount(t *testing.T) {
	store := &mockScoreAdminStore{
		getUserBalanceFn: func(ctx context.Context, uid int64) (int, error) { return 100, nil },
		saveScoreFn: func(ctx context.Context, uid int64, gt string, amount, balance int, reason string) error {
			if amount != -50 {
				t.Errorf("amount = %d, want -50", amount)
			}
			if balance != 50 {
				t.Errorf("balance = %d, want 50", balance)
			}
			return nil
		},
	}
	handler := NewScoreHandler(store)

	body := map[string]interface{}{
		"user_id": 1, "amount": -50, "reason": "penalty", "game_type": "admin",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/scores/adjust", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Adjust(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}
