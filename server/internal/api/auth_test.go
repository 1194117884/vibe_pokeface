package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yongkl/vibe-pokeface/internal/auth"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type mockUserStore struct {
	createFn         func(ctx context.Context, user *model.User) error
	findByProviderFn func(ctx context.Context, provider, providerUID string) (*model.User, error)
	createAuthFn     func(ctx context.Context, ua *model.UserAuth) error
	findAuthFn       func(ctx context.Context, provider, providerUID string) (*model.UserAuth, error)
	findByIDFn       func(ctx context.Context, id int64) (*model.User, error)
	listUsersFn      func(ctx context.Context, offset, limit int) ([]model.User, error)
	searchUsersFn    func(ctx context.Context, query string) ([]model.User, error)
	updateUserStatusFn func(ctx context.Context, userID int64, status int8) error
	getUserCountFn   func(ctx context.Context) (int, error)
}

func (m *mockUserStore) Create(ctx context.Context, user *model.User) error {
	return m.createFn(ctx, user)
}
func (m *mockUserStore) FindByProvider(ctx context.Context, provider, providerUID string) (*model.User, error) {
	return m.findByProviderFn(ctx, provider, providerUID)
}
func (m *mockUserStore) FindByID(ctx context.Context, id int64) (*model.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockUserStore) CreateAuth(ctx context.Context, ua *model.UserAuth) error {
	return m.createAuthFn(ctx, ua)
}
func (m *mockUserStore) FindAuth(ctx context.Context, provider, providerUID string) (*model.UserAuth, error) {
	return m.findAuthFn(ctx, provider, providerUID)
}
func (m *mockUserStore) ListUsers(ctx context.Context, offset, limit int) ([]model.User, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx, offset, limit)
	}
	return nil, nil
}
func (m *mockUserStore) SearchUsers(ctx context.Context, query string) ([]model.User, error) {
	if m.searchUsersFn != nil {
		return m.searchUsersFn(ctx, query)
	}
	return nil, nil
}
func (m *mockUserStore) UpdateUserStatus(ctx context.Context, userID int64, status int8) error {
	if m.updateUserStatusFn != nil {
		return m.updateUserStatusFn(ctx, userID, status)
	}
	return nil
}
func (m *mockUserStore) GetUserCount(ctx context.Context) (int, error) {
	if m.getUserCountFn != nil {
		return m.getUserCountFn(ctx)
	}
	return 0, nil
}

func TestRegister_Success(t *testing.T) {
	store := &mockUserStore{
		createFn: func(ctx context.Context, user *model.User) error {
			user.ID = 1
			return nil
		},
		findByProviderFn: func(ctx context.Context, provider, providerUID string) (*model.User, error) {
			return nil, nil
		},
		createAuthFn: func(ctx context.Context, ua *model.UserAuth) error {
			return nil
		},
		findAuthFn: func(ctx context.Context, provider, providerUID string) (*model.UserAuth, error) {
			return nil, nil
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc)

	body := map[string]string{"nickname": "testuser", "password": "secure123"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Error("response missing token")
	}
}

func TestRegister_MissingFields(t *testing.T) {
	store := &mockUserStore{}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc)

	body := map[string]string{"nickname": ""}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestLogin_Success(t *testing.T) {
	hashedPw, _ := auth.HashPassword("correct-pw")
	store := &mockUserStore{
		findAuthFn: func(ctx context.Context, provider, providerUID string) (*model.UserAuth, error) {
			return &model.UserAuth{UserID: 1, Credential: hashedPw}, nil
		},
		findByIDFn: func(ctx context.Context, id int64) (*model.User, error) {
			return &model.User{ID: 1, Nickname: "testuser", Role: "user"}, nil
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc)

	body := map[string]string{"password": "correct-pw"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Login(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hashedPw, _ := auth.HashPassword("correct-pw")
	store := &mockUserStore{
		findAuthFn: func(ctx context.Context, provider, providerUID string) (*model.UserAuth, error) {
			return &model.UserAuth{UserID: 1, Credential: hashedPw}, nil
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc)

	body := map[string]string{"password": "wrong-pw"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Login(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestGuestLogin_Success(t *testing.T) {
	store := &mockUserStore{
		findByProviderFn: func(ctx context.Context, provider, providerUID string) (*model.User, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, user *model.User) error {
			user.ID = 2
			return nil
		},
		createAuthFn: func(ctx context.Context, ua *model.UserAuth) error {
			return nil
		},
		findAuthFn: func(ctx context.Context, provider, providerUID string) (*model.UserAuth, error) {
			return nil, nil
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc)

	body := map[string]string{"device_id": "device-abc-123"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/guest", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.GuestLogin(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}
