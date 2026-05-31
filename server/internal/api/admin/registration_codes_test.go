package admin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/api/middleware"
	"github.com/yongkl/vibe-pokeface/internal/auth"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type mockRegCodeStore struct {
	createFn  func(ctx context.Context, codes []*model.RegistrationCode) error
	listFn    func(ctx context.Context, offset, limit int, usedFilter *bool) ([]model.RegistrationCodeWithUser, int, error)
	disableFn func(ctx context.Context, id int64) error
	countFn   func(ctx context.Context, used bool) (int, error)
}

func (m *mockRegCodeStore) Create(ctx context.Context, codes []*model.RegistrationCode) error {
	if m.createFn != nil {
		return m.createFn(ctx, codes)
	}
	return nil
}
func (m *mockRegCodeStore) List(ctx context.Context, offset, limit int, usedFilter *bool) ([]model.RegistrationCodeWithUser, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, offset, limit, usedFilter)
	}
	return nil, 0, nil
}
func (m *mockRegCodeStore) Disable(ctx context.Context, id int64) error {
	if m.disableFn != nil {
		return m.disableFn(ctx, id)
	}
	return nil
}
func (m *mockRegCodeStore) CountByStatus(ctx context.Context, used bool) (int, error) {
	if m.countFn != nil {
		return m.countFn(ctx, used)
	}
	return 0, nil
}

func TestGenerateCodes_Success(t *testing.T) {
	var createdCodes []*model.RegistrationCode
	store := &mockRegCodeStore{
		createFn: func(ctx context.Context, codes []*model.RegistrationCode) error {
			createdCodes = codes
			for i, c := range codes {
				c.ID = int64(i + 1)
				c.Code = "00000" + string(rune('1'+i))
			}
			return nil
		},
	}
	handler := NewRegistrationCodesHandler(store)

	body := map[string]interface{}{"count": 3, "note": "test batch"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/registration-codes", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.Claims{UserID: 1, Role: "admin"}
	ctx := context.WithValue(req.Context(), middleware.ClaimsKey, claims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.Generate(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d (201). Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}
	if len(createdCodes) != 3 {
		t.Errorf("created %d codes, want 3", len(createdCodes))
	}
}

func TestGenerateCodes_InvalidCount(t *testing.T) {
	store := &mockRegCodeStore{}
	handler := NewRegistrationCodesHandler(store)

	body := map[string]interface{}{"count": 0}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/registration-codes", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Generate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (400). Body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestGenerateCodes_MaxCount(t *testing.T) {
	store := &mockRegCodeStore{}
	handler := NewRegistrationCodesHandler(store)

	body := map[string]interface{}{"count": 101}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/registration-codes", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Generate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (400). Body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestListCodes_Success(t *testing.T) {
	store := &mockRegCodeStore{
		listFn: func(ctx context.Context, offset, limit int, usedFilter *bool) ([]model.RegistrationCodeWithUser, int, error) {
			return []model.RegistrationCodeWithUser{
				{RegistrationCode: model.RegistrationCode{ID: 1, Code: "123456"}},
			}, 1, nil
		},
	}
	handler := NewRegistrationCodesHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/registration-codes?page=1&size=20", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp codesListResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 1 {
		t.Errorf("total = %d, want 1", resp.Total)
	}
}

func TestDisableCode_Success(t *testing.T) {
	store := &mockRegCodeStore{
		disableFn: func(ctx context.Context, id int64) error {
			return nil
		},
	}
	handler := NewRegistrationCodesHandler(store)

	req := httptest.NewRequest("PUT", "/api/admin/registration-codes/1/disable", nil)
	rctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{"1"}},
	})
	req = req.WithContext(rctx)
	w := httptest.NewRecorder()
	handler.Disable(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestDisableCode_AlreadyUsed(t *testing.T) {
	store := &mockRegCodeStore{
		disableFn: func(ctx context.Context, id int64) error {
			return sql.ErrNoRows
		},
	}
	handler := NewRegistrationCodesHandler(store)

	req := httptest.NewRequest("PUT", "/api/admin/registration-codes/1/disable", nil)
	rctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{"1"}},
	})
	req = req.WithContext(rctx)
	w := httptest.NewRecorder()
	handler.Disable(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (400). Body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}
