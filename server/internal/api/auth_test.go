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
	createFn              func(ctx context.Context, user *model.User) error
	createUserWithAuthFn  func(ctx context.Context, user *model.User, auth *model.UserAuth) error
	findByProviderFn      func(ctx context.Context, provider, providerUID string) (*model.User, error)
	createAuthFn          func(ctx context.Context, ua *model.UserAuth) error
	findAuthFn            func(ctx context.Context, provider, providerUID string) (*model.UserAuth, error)
	findByIDFn            func(ctx context.Context, id int64) (*model.User, error)
	listUsersFn           func(ctx context.Context, offset, limit int) ([]model.User, error)
	searchUsersFn         func(ctx context.Context, query string) ([]model.User, error)
	updateUserStatusFn    func(ctx context.Context, userID int64, status int8) error
	getUserCountFn        func(ctx context.Context) (int, error)
}

func (m *mockUserStore) Create(ctx context.Context, user *model.User) error {
	return m.createFn(ctx, user)
}
func (m *mockUserStore) CreateUserWithAuth(ctx context.Context, user *model.User, auth *model.UserAuth) error {
	if m.createUserWithAuthFn != nil {
		return m.createUserWithAuthFn(ctx, user, auth)
	}
	return nil
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
		createUserWithAuthFn: func(ctx context.Context, user *model.User, auth *model.UserAuth) error {
			user.ID = 1
			auth.UserID = 1
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

	body := map[string]string{"password": "correct-pw", "provider_uid": "password:testuser"}
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

	body := map[string]string{"password": "wrong-pw", "provider_uid": "password:testuser"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Login(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRegister_DuplicateNickname(t *testing.T) {
	store := &mockUserStore{
		createUserWithAuthFn: func(ctx context.Context, user *model.User, auth *model.UserAuth) error {
			return model.ErrDuplicateNickname
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc)

	body := map[string]string{"nickname": "existing", "password": "secure123"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d (409). Body: %s", w.Code, http.StatusConflict, w.Body.String())
	}
}

func TestRegister_DuplicateAuth(t *testing.T) {
	store := &mockUserStore{
		createUserWithAuthFn: func(ctx context.Context, user *model.User, auth *model.UserAuth) error {
			return model.ErrDuplicateAuth
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc)

	body := map[string]string{"nickname": "existing", "password": "secure123"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d (409). Body: %s", w.Code, http.StatusConflict, w.Body.String())
	}
}

func TestGuestLogin_RaceCondition(t *testing.T) {
	store := &mockUserStore{
		findByProviderFn: func(ctx context.Context, provider, providerUID string) (*model.User, error) {
			return nil, nil
		},
		createUserWithAuthFn: func(ctx context.Context, user *model.User, auth *model.UserAuth) error {
			return model.ErrDuplicateAuth
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc)

	body := map[string]string{"device_id": "race-device"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/guest", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.GuestLogin(w, req)

	// Without a FindByProvider recovery, this would be a server error.
	// Since FindByProvider returns nil too, the handler should return a
	// proper error rather than crashing.
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d (500). Body: %s", w.Code, http.StatusInternalServerError, w.Body.String())
	}
}

func TestGuestLogin_Success(t *testing.T) {
	store := &mockUserStore{
		findByProviderFn: func(ctx context.Context, provider, providerUID string) (*model.User, error) {
			return nil, nil
		},
		createUserWithAuthFn: func(ctx context.Context, user *model.User, auth *model.UserAuth) error {
			user.ID = 2
			auth.UserID = 2
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
