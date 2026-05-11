package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type mockUserAdminStore struct {
	listUsersFn         func(ctx context.Context, offset, limit int) ([]model.User, error)
	searchUsersFn       func(ctx context.Context, query string) ([]model.User, error)
	updateUserStatusFn  func(ctx context.Context, userID int64, status int8) error
	getUserCountFn      func(ctx context.Context) (int, error)
	findByIDFn          func(ctx context.Context, id int64) (*model.User, error)
}

func (m *mockUserAdminStore) ListUsers(ctx context.Context, offset, limit int) ([]model.User, error) {
	return m.listUsersFn(ctx, offset, limit)
}
func (m *mockUserAdminStore) SearchUsers(ctx context.Context, query string) ([]model.User, error) {
	return m.searchUsersFn(ctx, query)
}
func (m *mockUserAdminStore) UpdateUserStatus(ctx context.Context, userID int64, status int8) error {
	return m.updateUserStatusFn(ctx, userID, status)
}
func (m *mockUserAdminStore) GetUserCount(ctx context.Context) (int, error) {
	return m.getUserCountFn(ctx)
}
func (m *mockUserAdminStore) FindByID(ctx context.Context, id int64) (*model.User, error) {
	return m.findByIDFn(ctx, id)
}

func TestUsersList(t *testing.T) {
	users := []model.User{
		{ID: 1, Nickname: "Alice", Role: "user", Status: 1},
		{ID: 2, Nickname: "Bob", Role: "admin", Status: 1},
	}
	store := &mockUserAdminStore{
		listUsersFn: func(ctx context.Context, offset, limit int) ([]model.User, error) { return users, nil },
		getUserCountFn: func(ctx context.Context) (int, error) { return 10, nil },
	}
	handler := NewAdminUserHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/users?page=1&size=20", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp userListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Users) != 2 {
		t.Errorf("users count = %d, want 2", len(resp.Users))
	}
	if resp.Total != 10 {
		t.Errorf("total = %d, want 10", resp.Total)
	}
}

func TestUsersList_Empty(t *testing.T) {
	store := &mockUserAdminStore{
		listUsersFn:    func(ctx context.Context, offset, limit int) ([]model.User, error) { return []model.User{}, nil },
		getUserCountFn: func(ctx context.Context) (int, error) { return 0, nil },
	}
	handler := NewAdminUserHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/users", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var resp userListResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Users) != 0 {
		t.Errorf("users = %v, want empty", resp.Users)
	}
}

func TestUsersSearch(t *testing.T) {
	store := &mockUserAdminStore{
		searchUsersFn: func(ctx context.Context, q string) ([]model.User, error) {
			if q != "Ali" {
				t.Errorf("query = %q, want %q", q, "Ali")
			}
			return []model.User{{ID: 1, Nickname: "Alice", Role: "user"}}, nil
		},
	}
	handler := NewAdminUserHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/users?q=Ali", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var resp userListResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Users) != 1 || resp.Users[0].Nickname != "Alice" {
		t.Errorf("unexpected results: %+v", resp.Users)
	}
}

func TestUsersUpdateStatus_Success(t *testing.T) {
	store := &mockUserAdminStore{
		updateUserStatusFn: func(ctx context.Context, userID int64, status int8) error {
			if userID != 1 {
				t.Errorf("userID = %d, want 1", userID)
			}
			if status != 0 {
				t.Errorf("status = %d, want 0", status)
			}
			return nil
		},
	}
	handler := NewAdminUserHandler(store)

	body := `{"status":0}`
	req := httptest.NewRequest("PUT", "/api/admin/users/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set chi URL param context
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	w := httptest.NewRecorder()
	handler.UpdateStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("response = %v, want %v", resp, map[string]string{"status": "ok"})
	}
}

func TestUsersUpdateStatus_InvalidID(t *testing.T) {
	handler := NewAdminUserHandler(&mockUserAdminStore{})

	body := `{"status":0}`
	req := httptest.NewRequest("PUT", "/api/admin/users/abc/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set invalid id chi URL param
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	w := httptest.NewRecorder()
	handler.UpdateStatus(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersUpdateStatus_InvalidBody(t *testing.T) {
	handler := NewAdminUserHandler(&mockUserAdminStore{})

	body := `not-json`
	req := httptest.NewRequest("PUT", "/api/admin/users/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	w := httptest.NewRecorder()
	handler.UpdateStatus(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersList_DefaultPagination(t *testing.T) {
	store := &mockUserAdminStore{
		listUsersFn: func(ctx context.Context, offset, limit int) ([]model.User, error) {
			if offset != 0 {
				t.Errorf("offset = %d, want 0", offset)
			}
			if limit != 20 {
				t.Errorf("limit = %d, want 20", limit)
			}
			return nil, nil
		},
		getUserCountFn: func(ctx context.Context) (int, error) { return 0, nil },
	}
	handler := NewAdminUserHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/users", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}
