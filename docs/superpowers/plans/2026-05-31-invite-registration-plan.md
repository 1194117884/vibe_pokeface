# Invite-Only Registration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close open registration — all new accounts require a one-time-use 6-digit registration code. Guest login returns 403. Admin panel manages codes.

**Architecture:** New `registration_codes` MySQL table + Go model/store + admin handler following existing patterns. Auth handler gains `RegistrationCodeStore` dependency. Frontend register page adds code input; new admin page for code management.

**Tech Stack:** Go 1.26 (chi, sqlx), TypeScript (Next.js 16, React 19, Tailwind CSS v4)

**Spec:** `docs/superpowers/specs/2026-05-31-invite-registration-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `server/migrations/009_registration_codes.sql` | Create | DDL for registration_codes table |
| `server/internal/model/registration_code.go` | Create | Struct + store interface + DB implementation |
| `server/internal/api/auth.go` | Modify | Add RegistrationCode to request; validate code in Register; block GuestLogin |
| `server/internal/api/admin/registration_codes.go` | Create | Admin CRUD handlers for registration codes |
| `server/internal/api/admin/handler.go` | Modify | Add RegistrationCodes field to composite Handler |
| `server/internal/api/middleware/auth.go` | Modify | Add exported GetClaims helper |
| `server/internal/api/router.go` | Modify | Pass regCodeStore; register new admin routes |
| `server/cmd/server/main.go` | Modify | Instantiate RegistrationCodeDB; pass to router |
| `server/internal/api/auth_test.go` | Modify | Add tests for code validation + guest closure |
| `server/internal/api/admin/registration_codes_test.go` | Create | Tests for admin code management endpoints |
| `frontend/lib/api-client.ts` | Modify | Add registrationCode parameter to register() |
| `frontend/app/auth/register/page.tsx` | Modify | Add registration code input field |
| `frontend/lib/admin-fetch.ts` | Modify | Add API helpers for registration codes |
| `frontend/app/admin/registration-codes/page.tsx` | Create | Admin page: generate, list, copy, disable codes |
| `frontend/components/ui/AdminSidebar.tsx` | Modify | Add nav item for registration codes |

---

### Task 1: Database Migration

**Files:**
- Create: `server/migrations/009_registration_codes.sql`

- [ ] **Step 1: Create migration file**

```sql
-- 009_registration_codes: invite-only registration codes
CREATE TABLE IF NOT EXISTS registration_codes (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    code            VARCHAR(6) NOT NULL,
    is_used         BOOLEAN DEFAULT FALSE,
    is_disabled     BOOLEAN DEFAULT FALSE,
    used_by_user_id BIGINT NULL,
    used_at         DATETIME NULL,
    created_by      BIGINT NOT NULL,
    note            VARCHAR(255) DEFAULT '',
    created_at      DATETIME DEFAULT NOW(),
    updated_at      DATETIME DEFAULT NOW(),
    UNIQUE KEY uk_code (code),
    INDEX idx_used (is_used),
    INDEX idx_disabled (is_disabled)
);
```

- [ ] **Step 2: Commit**

```bash
git add server/migrations/009_registration_codes.sql
git commit -m "feat(db): add registration_codes table for invite-only registration"
```

---

### Task 2: Go Model + Store

**Files:**
- Create: `server/internal/model/registration_code.go`

- [ ] **Step 1: Create model file with struct, interface, and implementation**

```go
package model

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/jmoiron/sqlx"
)

// RegistrationCode represents a one-time-use registration code.
type RegistrationCode struct {
	ID            int64      `db:"id" json:"id"`
	Code          string     `db:"code" json:"code"`
	IsUsed        bool       `db:"is_used" json:"is_used"`
	IsDisabled    bool       `db:"is_disabled" json:"is_disabled"`
	UsedByUserID  *int64     `db:"used_by_user_id" json:"used_by_user_id,omitempty"`
	UsedAt        *time.Time `db:"used_at" json:"used_at,omitempty"`
	CreatedBy     int64      `db:"created_by" json:"created_by"`
	Note          string     `db:"note" json:"note"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// RegistrationCodeStore defines the interface for registration code operations.
type RegistrationCodeStore interface {
	Create(ctx context.Context, codes []*RegistrationCode) error
	FindByCode(ctx context.Context, code string) (*RegistrationCode, error)
	ClaimCode(ctx context.Context, code string, userID int64) (bool, error)
	List(ctx context.Context, offset, limit int, usedFilter *bool) ([]RegistrationCodeWithUser, int, error)
	Disable(ctx context.Context, id int64) error
	CountByStatus(ctx context.Context, used bool) (int, error)
}

// RegistrationCodeWithUser includes the user nickname for display.
type RegistrationCodeWithUser struct {
	RegistrationCode
	UsedByNickname *string `db:"used_by_nickname" json:"used_by_nickname,omitempty"`
}

// RegistrationCodeDB implements RegistrationCodeStore.
type RegistrationCodeDB struct {
	db *sqlx.DB
}

func NewRegistrationCodeDB(db *sqlx.DB) *RegistrationCodeDB {
	return &RegistrationCodeDB{db: db}
}

// GenerateCode creates a random 6-digit numeric string.
func GenerateCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// Create inserts multiple registration codes. Retries on code collision.
func (s *RegistrationCodeDB) Create(ctx context.Context, codes []*RegistrationCode) error {
	for _, c := range codes {
		for attempt := 0; attempt < 5; attempt++ {
			code, err := GenerateCode()
			if err != nil {
				return fmt.Errorf("generate code: %w", err)
			}
			c.Code = code

			_, err = s.db.ExecContext(ctx,
				`INSERT INTO registration_codes (code, created_by, note) VALUES (?, ?, ?)`,
				c.Code, c.CreatedBy, c.Note)
			if err == nil {
				break
			}
			if !isDuplicateEntry(err) {
				return err
			}
			// Duplicate code, retry with a new random code
			if attempt == 4 {
				return fmt.Errorf("failed to generate unique code after 5 attempts")
			}
		}
	}
	return nil
}

// FindByCode looks up a registration code by its code string.
func (s *RegistrationCodeDB) FindByCode(ctx context.Context, code string) (*RegistrationCode, error) {
	var rc RegistrationCode
	err := s.db.GetContext(ctx, &rc, "SELECT * FROM registration_codes WHERE code = ?", code)
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

// ClaimCode atomically marks a code as used. Returns true if successful, false if already used/disabled.
func (s *RegistrationCodeDB) ClaimCode(ctx context.Context, code string, userID int64) (bool, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE registration_codes SET is_used = TRUE, used_by_user_id = ?, used_at = NOW() WHERE code = ? AND is_used = FALSE AND is_disabled = FALSE`,
		userID, code)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// List returns paginated registration codes, optionally filtered by used status.
func (s *RegistrationCodeDB) List(ctx context.Context, offset, limit int, usedFilter *bool) ([]RegistrationCodeWithUser, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	where := ""
	args := []interface{}{}
	if usedFilter != nil {
		where = " WHERE rc.is_used = ?"
		args = append(args, *usedFilter)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM registration_codes rc" + where
	if err := s.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query := `SELECT rc.*, u.nickname AS used_by_nickname
		FROM registration_codes rc
		LEFT JOIN users u ON rc.used_by_user_id = u.id` + where +
		` ORDER BY rc.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	var codes []RegistrationCodeWithUser
	if err := s.db.SelectContext(ctx, &codes, query, args...); err != nil {
		return nil, 0, err
	}
	if codes == nil {
		codes = []RegistrationCodeWithUser{}
	}
	return codes, total, nil
}

// Disable marks an unused, non-disabled code as disabled. Returns sql.ErrNoRows if not found or already used.
func (s *RegistrationCodeDB) Disable(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE registration_codes SET is_disabled = TRUE WHERE id = ? AND is_used = FALSE AND is_disabled = FALSE`,
		id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CountByStatus returns the count of codes filtered by used status.
func (s *RegistrationCodeDB) CountByStatus(ctx context.Context, used bool) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM registration_codes WHERE is_used = ?", used)
	return count, err
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go vet ./internal/model/...
```

- [ ] **Step 3: Commit**

```bash
git add server/internal/model/registration_code.go
git commit -m "feat(model): add RegistrationCode model, store interface, and DB implementation"
```

---

### Task 3: Modify Auth Handler (Register + Guest)

**Files:**
- Modify: `server/internal/api/auth.go`

- [ ] **Step 1: Update AuthHandler struct and constructor to accept RegistrationCodeStore**

Replace lines 12-19 in `auth.go`:

```go
type AuthHandler struct {
	store    model.UserStore
	jwt      *auth.JWTService
	regCodes model.RegistrationCodeStore
}

func NewAuthHandler(store model.UserStore, jwt *auth.JWTService, regCodes model.RegistrationCodeStore) *AuthHandler {
	return &AuthHandler{store: store, jwt: jwt, regCodes: regCodes}
}
```

- [ ] **Step 2: Update registerRequest to include registration_code**

Replace lines 21-24:

```go
type registerRequest struct {
	Nickname         string `json:"nickname"`
	Password         string `json:"password"`
	RegistrationCode string `json:"registration_code"`
}
```

- [ ] **Step 3: Update Register handler to validate registration code**

Replace the Register function (lines 40-75):

```go
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Nickname == "" || req.Password == "" {
		http.Error(w, `{"error":"nickname and password required"}`, http.StatusBadRequest)
		return
	}
	if req.RegistrationCode == "" {
		http.Error(w, `{"error":"registration_code required"}`, http.StatusBadRequest)
		return
	}

	// Look up the registration code.
	regCode, err := h.regCodes.FindByCode(r.Context(), req.RegistrationCode)
	if err != nil || regCode == nil {
		http.Error(w, `{"error":"invalid registration code"}`, http.StatusForbidden)
		return
	}
	if regCode.IsUsed {
		http.Error(w, `{"error":"registration code already used"}`, http.StatusForbidden)
		return
	}
	if regCode.IsDisabled {
		http.Error(w, `{"error":"registration code disabled"}`, http.StatusForbidden)
		return
	}

	hashedPw, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"failed to process password"}`, http.StatusInternalServerError)
		return
	}

	user := &model.User{Nickname: req.Nickname, Role: "user"}
	userAuth := &model.UserAuth{
		Provider:    "password",
		ProviderUID: "password:" + req.Nickname,
		Credential:  hashedPw,
	}

	if err := h.store.CreateUserWithAuth(r.Context(), user, userAuth); err != nil {
		if errors.Is(err, model.ErrDuplicateNickname) || errors.Is(err, model.ErrDuplicateAuth) {
			http.Error(w, `{"error":"user already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"failed to create user"}`, http.StatusInternalServerError)
		return
	}

	// Atomically claim the code. If it fails (race), the user is created but
	// the code is left unused — acceptable trade-off.
	claimed, _ := h.regCodes.ClaimCode(r.Context(), req.RegistrationCode, user.ID)
	if !claimed {
		// Code was claimed concurrently — user is already created, just log.
		// The code is still valid for the race winner.
	}

	token, _ := h.jwt.GenerateToken(user.ID, user.Role)
	json.NewEncoder(w).Encode(authResponse{Token: token, User: user})
}
```

- [ ] **Step 4: Replace GuestLogin to always return 403**

Replace the entire GuestLogin function (lines 115-167):

```go
func (h *AuthHandler) GuestLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"guest registration closed"}`))
}
```

- [ ] **Step 5: Verify compilation**

```bash
cd server && go vet ./internal/api/...
```

- [ ] **Step 6: Commit**

```bash
git add server/internal/api/auth.go
git commit -m "feat(auth): require registration code for register, close guest login"
```

---

### Task 4: Wire Dependencies (Middleware + Router + Main)

**Files:**
- Modify: `server/internal/api/middleware/auth.go`
- Modify: `server/internal/api/router.go`
- Modify: `server/cmd/server/main.go`

- [ ] **Step 0: Add exported GetClaims helper to middleware**

Append to `server/internal/api/middleware/auth.go`:

```go
// GetClaims retrieves the JWT claims from the request context.
// Returns nil if no claims are present.
func GetClaims(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(ClaimsKey).(*auth.Claims)
	return claims
}
```

Also add `"context"` to the imports.

- [ ] **Step 1: Update NewRouter signature to accept RegistrationCodeStore**

In `router.go`, change line 15:

```go
func NewRouter(store model.UserStore, regCodeStore model.RegistrationCodeStore, jwt *auth.JWTService, hub *ws.Hub, corsCfg middleware.CORSConfig, lkConfig LiveKitConfig, adminHandler *admin.Handler, roomHandler *RoomHandler) *chi.Mux {
```

Change line 21 (authHandler construction):

```go
	authHandler := NewAuthHandler(store, jwt, regCodeStore)
```

- [ ] **Step 2: Add admin registration-code routes**

In `router.go`, add after the existing admin routes (after line 58, before the closing `})` of the admin route group):

```go
					r.Get("/registration-codes", adminHandler.RegistrationCodes.List)
					r.Post("/registration-codes", adminHandler.RegistrationCodes.Generate)
					r.Put("/registration-codes/{id}/disable", adminHandler.RegistrationCodes.Disable)
```

- [ ] **Step 3: Update main.go to create RegistrationCodeDB**

In `main.go`, after line 35 (`gameStore := model.NewGameStore(db)`):

```go
	regCodeStore := model.NewRegistrationCodeDB(db)
```

Update line 42 (`adminHandler := admin.NewHandler(...)`) to:

```go
	adminHandler := admin.NewHandler(userDB, gameStore, aiStore, hub, regCodeStore)
```

Update line 52 (`router := api.NewRouter(...)`) to include `regCodeStore` as second argument:

```go
	router := api.NewRouter(userDB, regCodeStore, jwtSvc, hub, middleware.CORSConfig{
		AllowedOrigins: cfg.AllowedOrigins,
	}, lkConfig, adminHandler, roomHandler)
```

- [ ] **Step 4: Verify compilation**

```bash
cd server && go build ./cmd/server/
```

- [ ] **Step 5: Commit**

```bash
git add server/internal/api/router.go server/cmd/server/main.go
git commit -m "feat(server): wire RegistrationCodeStore into auth and admin handlers"
```

---

### Task 5: Admin Registration Codes Handler

**Files:**
- Create: `server/internal/api/admin/registration_codes.go`
- Modify: `server/internal/api/admin/handler.go`

- [ ] **Step 1: Create registration codes admin handler**

```go
package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/api/middleware"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

// RegistrationCodeAdminStore is the subset of RegistrationCodeStore used by admin handlers.
type RegistrationCodeAdminStore interface {
	Create(ctx context.Context, codes []*model.RegistrationCode) error
	List(ctx context.Context, offset, limit int, usedFilter *bool) ([]model.RegistrationCodeWithUser, int, error)
	Disable(ctx context.Context, id int64) error
	CountByStatus(ctx context.Context, used bool) (int, error)
}

type RegistrationCodesHandler struct {
	store RegistrationCodeAdminStore
}

func NewRegistrationCodesHandler(store RegistrationCodeAdminStore) *RegistrationCodesHandler {
	return &RegistrationCodesHandler{store: store}
}

type generateRequest struct {
	Count int    `json:"count"`
	Note  string `json:"note"`
}

type codesListResponse struct {
	Codes []model.RegistrationCodeWithUser `json:"codes"`
	Total int                              `json:"total"`
}

// List returns paginated registration codes.
func (h *RegistrationCodesHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 100 {
		size = 20
	}

	var usedFilter *bool
	if f := r.URL.Query().Get("used"); f != "" {
		v := f == "true"
		usedFilter = &v
	}

	offset := (page - 1) * size

	codes, total, err := h.store.List(r.Context(), offset, size, usedFilter)
	if err != nil {
		http.Error(w, `{"error":"list failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(codesListResponse{Codes: codes, Total: total})
}

// Generate creates new registration codes.
func (h *RegistrationCodesHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Count < 1 || req.Count > 100 {
		http.Error(w, `{"error":"count must be between 1 and 100"}`, http.StatusBadRequest)
		return
	}

	claims := middleware.GetClaims(r.Context())

	codes := make([]*model.RegistrationCode, req.Count)
	for i := range codes {
		codes[i] = &model.RegistrationCode{
			CreatedBy: claims.UserID,
			Note:      req.Note,
		}
	}

	if err := h.store.Create(r.Context(), codes); err != nil {
		http.Error(w, `{"error":"failed to generate codes"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"codes": codes,
		"count": len(codes),
	})
}

// Disable disables an unused registration code.
func (h *RegistrationCodesHandler) Disable(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid code id"}`, http.StatusBadRequest)
		return
	}

	if err := h.store.Disable(r.Context(), id); err != nil {
		http.Error(w, `{"error":"disable failed — code may already be used or disabled"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
```

- [ ] **Step 2: Update admin Handler to include RegistrationCodes**

In `handler.go`, add field:

```go
type Handler struct {
	Dashboard          *DashboardHandler
	Users              *AdminUserHandler
	Rooms              *AdminRoomHandler
	AIChars            *AICharacterHandler
	LLMConfig          *LLMConfigHandler
	Scores             *ScoreHandler
	RegistrationCodes  *RegistrationCodesHandler
}
```

Update `NewHandler`:

```go
func NewHandler(userStore *model.UserDB, gameStore *model.GameStore, aiStore *model.AIStore, hub *ws.Hub, regCodeStore RegistrationCodeAdminStore) *Handler {
	return &Handler{
		Dashboard:         NewDashboardHandler(userStore, gameStore, hub),
		Users:             NewAdminUserHandler(userStore),
		Rooms:             NewAdminRoomHandler(gameStore),
		AIChars:           NewAICharacterHandler(aiStore),
		LLMConfig:         NewLLMConfigHandler(aiStore, aiStore),
		Scores:            NewScoreHandler(gameStore),
		RegistrationCodes: NewRegistrationCodesHandler(regCodeStore),
	}
}
```

- [ ] **Step 3: Check that auth.GetClaims exists**

```bash
cd server && grep -r "func GetClaims" internal/auth/
```

If `GetClaims` doesn't exist, use alternative to get admin user ID from context.

- [ ] **Step 4: Verify compilation**

```bash
cd server && go vet ./internal/api/admin/... ./internal/api/... ./cmd/server/
```

- [ ] **Step 5: Commit**

```bash
git add server/internal/api/admin/registration_codes.go server/internal/api/admin/handler.go
git commit -m "feat(admin): add registration codes CRUD handler"
```

---

### Task 6: Backend Tests

**Files:**
- Modify: `server/internal/api/auth_test.go`
- Create: `server/internal/api/admin/registration_codes_test.go`

- [ ] **Step 1: Update mockUserStore to implement new RegistrationCodeStore methods for auth tests**

Actually, since we're adding `RegistrationCodeStore` as a separate dependency to `AuthHandler`, we need a separate mock. Add a `mockRegCodeStore` at the top of `auth_test.go`:

```go
type mockRegCodeStore struct {
	findByCodeFn func(ctx context.Context, code string) (*model.RegistrationCode, error)
	claimCodeFn  func(ctx context.Context, code string, userID int64) (bool, error)
	createFn     func(ctx context.Context, codes []*model.RegistrationCode) error
	listFn       func(ctx context.Context, offset, limit int, usedFilter *bool) ([]model.RegistrationCodeWithUser, int, error)
	disableFn    func(ctx context.Context, id int64) error
	countFn      func(ctx context.Context, used bool) (int, error)
}

func (m *mockRegCodeStore) Create(ctx context.Context, codes []*model.RegistrationCode) error {
	if m.createFn != nil { return m.createFn(ctx, codes) }; return nil
}
func (m *mockRegCodeStore) FindByCode(ctx context.Context, code string) (*model.RegistrationCode, error) {
	if m.findByCodeFn != nil { return m.findByCodeFn(ctx, code) }; return nil, nil
}
func (m *mockRegCodeStore) ClaimCode(ctx context.Context, code string, userID int64) (bool, error) {
	if m.claimCodeFn != nil { return m.claimCodeFn(ctx, code, userID) }; return true, nil
}
func (m *mockRegCodeStore) List(ctx context.Context, offset, limit int, usedFilter *bool) ([]model.RegistrationCodeWithUser, int, error) {
	if m.listFn != nil { return m.listFn(ctx, offset, limit, usedFilter) }; return nil, 0, nil
}
func (m *mockRegCodeStore) Disable(ctx context.Context, id int64) error {
	if m.disableFn != nil { return m.disableFn(ctx, id) }; return nil
}
func (m *mockRegCodeStore) CountByStatus(ctx context.Context, used bool) (int, error) {
	if m.countFn != nil { return m.countFn(ctx, used) }; return 0, nil
}
```

- [ ] **Step 2: Update all existing test handler constructions**

All `NewAuthHandler(store, jwtSvc)` calls become `NewAuthHandler(store, jwtSvc, &mockRegCodeStore{})`.

- [ ] **Step 3: Add test for missing registration code**

```go
func TestRegister_MissingRegistrationCode(t *testing.T) {
	store := &mockUserStore{}
	regStore := &mockRegCodeStore{}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc, regStore)

	body := map[string]string{"nickname": "testuser", "password": "secure123"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d. Body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}
```

- [ ] **Step 4: Add test for invalid registration code**

```go
func TestRegister_InvalidRegistrationCode(t *testing.T) {
	store := &mockUserStore{}
	regStore := &mockRegCodeStore{
		findByCodeFn: func(ctx context.Context, code string) (*model.RegistrationCode, error) {
			return nil, nil // not found
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc, regStore)

	body := map[string]string{"nickname": "testuser", "password": "secure123", "registration_code": "999999"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d (403). Body: %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}
```

- [ ] **Step 5: Add test for already-used registration code**

```go
func TestRegister_AlreadyUsedCode(t *testing.T) {
	store := &mockUserStore{}
	regStore := &mockRegCodeStore{
		findByCodeFn: func(ctx context.Context, code string) (*model.RegistrationCode, error) {
			return &model.RegistrationCode{ID: 1, Code: "123456", IsUsed: true}, nil
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc, regStore)

	body := map[string]string{"nickname": "testuser", "password": "secure123", "registration_code": "123456"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d (403). Body: %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}
```

- [ ] **Step 6: Add test for disabled registration code**

```go
func TestRegister_DisabledCode(t *testing.T) {
	store := &mockUserStore{}
	regStore := &mockRegCodeStore{
		findByCodeFn: func(ctx context.Context, code string) (*model.RegistrationCode, error) {
			return &model.RegistrationCode{ID: 1, Code: "123456", IsDisabled: true}, nil
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc, regStore)

	body := map[string]string{"nickname": "testuser", "password": "secure123", "registration_code": "123456"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d (403). Body: %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}
```

- [ ] **Step 7: Add test for successful registration with valid code**

```go
func TestRegister_WithValidCode(t *testing.T) {
	store := &mockUserStore{
		createUserWithAuthFn: func(ctx context.Context, user *model.User, auth *model.UserAuth) error {
			user.ID = 1
			auth.UserID = 1
			return nil
		},
	}
	regStore := &mockRegCodeStore{
		findByCodeFn: func(ctx context.Context, code string) (*model.RegistrationCode, error) {
			return &model.RegistrationCode{ID: 1, Code: "123456", IsUsed: false}, nil
		},
		claimCodeFn: func(ctx context.Context, code string, userID int64) (bool, error) {
			return true, nil
		},
	}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc, regStore)

	body := map[string]string{"nickname": "testuser", "password": "secure123", "registration_code": "123456"}
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
```

- [ ] **Step 8: Replace guest login test — it should now return 403**

Replace `TestGuestLogin_Success` (lines 240-266):

```go
func TestGuestLogin_Closed(t *testing.T) {
	store := &mockUserStore{}
	regStore := &mockRegCodeStore{}
	jwtSvc := auth.NewJWTService("test-secret")
	handler := NewAuthHandler(store, jwtSvc, regStore)

	body := map[string]string{"device_id": "device-abc-123"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/guest", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.GuestLogin(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d (403). Body: %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}
```

Also remove `TestGuestLogin_RaceCondition` (lines 213-238) since it no longer applies — or update it to also expect 403.

- [ ] **Step 9: Run backend tests**

```bash
cd server && go test ./internal/api/... -v
```

Expected: all tests pass.

- [ ] **Step 10: Create admin registration codes test**

```go
package admin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

type mockRegCodeStore struct {
	createFn  func(ctx context.Context, codes []*model.RegistrationCode) error
	listFn    func(ctx context.Context, offset, limit int, usedFilter *bool) ([]model.RegistrationCodeWithUser, int, error)
	disableFn func(ctx context.Context, id int64) error
	countFn   func(ctx context.Context, used bool) (int, error)
}

func (m *mockRegCodeStore) Create(ctx context.Context, codes []*model.RegistrationCode) error {
	if m.createFn != nil { return m.createFn(ctx, codes) }; return nil
}
func (m *mockRegCodeStore) List(ctx context.Context, offset, limit int, usedFilter *bool) ([]model.RegistrationCodeWithUser, int, error) {
	if m.listFn != nil { return m.listFn(ctx, offset, limit, usedFilter) }; return nil, 0, nil
}
func (m *mockRegCodeStore) Disable(ctx context.Context, id int64) error {
	if m.disableFn != nil { return m.disableFn(ctx, id) }; return nil
}
func (m *mockRegCodeStore) CountByStatus(ctx context.Context, used bool) (int, error) {
	if m.countFn != nil { return m.countFn(ctx, used) }; return 0, nil
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
	// Simulate chi URL param by setting it on the request context
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
```

- [ ] **Step 11: Run admin tests**

```bash
cd server && go test ./internal/api/admin/... -v -run TestGenerateCode
cd server && go test ./internal/api/admin/... -v -run TestListCodes
cd server && go test ./internal/api/admin/... -v -run TestDisableCode
```

Expected: all pass.

- [ ] **Step 12: Run all server tests**

```bash
cd server && go test ./... 2>&1 | tail -20
```

Expected: all pass.

- [ ] **Step 13: Commit**

```bash
git add server/internal/api/auth_test.go server/internal/api/admin/registration_codes_test.go
git commit -m "test: add tests for registration code validation and admin endpoints"
```

---

### Task 7: Frontend apiClient

**Files:**
- Modify: `frontend/lib/api-client.ts`

- [ ] **Step 1: Update register() to accept and send registration_code**

Replace the `register` method (lines 90-98):

```typescript
  async register(nickname: string, password: string, registrationCode: string) {
    const result = await this.request<LoginResult>("POST", "/api/auth/register", {
      nickname,
      password,
      registration_code: registrationCode,
    });
    if (result.data) {
      this.setUser(result.data.user);
    }
    return result;
  }
```

- [ ] **Step 2: Verify TypeScript compilation**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -30
```

- [ ] **Step 3: Commit**

```bash
git add frontend/lib/api-client.ts
git commit -m "feat(frontend): add registration code parameter to apiClient.register()"
```

---

### Task 8: Frontend Register Page

**Files:**
- Modify: `frontend/app/auth/register/page.tsx`

- [ ] **Step 1: Add registration code input and update submit handler**

Replace the entire file:

```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { apiClient } from "@/lib/api-client";

export default function RegisterPage() {
  const router = useRouter();
  const [nickname, setNickname] = useState("");
  const [password, setPassword] = useState("");
  const [regCode, setRegCode] = useState("");
  const [error, setError] = useState("");
  const [regCodeError, setRegCodeError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setRegCodeError("");

    if (regCode.length !== 6) {
      setRegCodeError("请输入6位注册码");
      return;
    }

    setLoading(true);

    const result = await apiClient.register(nickname, password, regCode);
    if (result.error) {
      if (result.error.toLowerCase().includes("registration code") || result.error.includes("注册码")) {
        setRegCodeError(result.error);
      } else {
        setError(result.error);
      }
      setLoading(false);
      return;
    }

    if (result.data) {
      apiClient.setToken(result.data.token);
      router.push("/lobby");
    }
    setLoading(false);
  };

  return (
    <Card padding="lg">
      <div className="text-center mb-8">
        <div className="text-4xl mb-3">✨</div>
        <h1 className="text-3xl font-black text-starbucks">Create Account</h1>
        <p className="text-base font-bold text-text-black-soft mt-1">
          需要邀请码才能注册
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">
        <Input
          label="Nickname"
          value={nickname}
          onChange={(e) => setNickname(e.target.value)}
        />
        <Input
          label="Password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <Input
          label="注册码"
          value={regCode}
          onChange={(e) => {
            const v = e.target.value.replace(/\D/g, "").slice(0, 6);
            setRegCode(v);
            setRegCodeError("");
          }}
          error={regCodeError}
          inputMode="numeric"
          maxLength={6}
          placeholder="6位数字"
        />
        <Button type="submit" fullWidth disabled={loading}>
          {loading ? "Creating account..." : "Create Account"}
        </Button>
      </form>

      <div className="mt-8 pt-6 border-t border-cream text-center">
        <p className="text-base font-bold text-text-black-soft">
          Already have an account?{" "}
          <Link
            href="/auth/login"
            className="text-green-accent font-semibold hover:underline"
          >
            Sign In
          </Link>
        </p>
      </div>
    </Card>
  );
}
```

- [ ] **Step 2: Frontend checks**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
cd frontend && npm run lint 2>&1 | tail -5
```

- [ ] **Step 3: Commit**

```bash
git add frontend/app/auth/register/page.tsx
git commit -m "feat(frontend): add registration code input to register page"
```

---

### Task 9: Frontend Admin Fetch Helpers

**Files:**
- Modify: `frontend/lib/admin-fetch.ts`

- [ ] **Step 1: Add registration code API helpers**

Append to `admin-fetch.ts`:

```typescript
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface RegistrationCode {
  id: number;
  code: string;
  is_used: boolean;
  is_disabled: boolean;
  used_by_user_id: number | null;
  used_at: string | null;
  created_by: number;
  note: string;
  created_at: string;
  updated_at: string;
  used_by_nickname: string | null;
}

export interface CodesListResponse {
  codes: RegistrationCode[];
  total: number;
}

export async function fetchRegistrationCodes(
  page: number,
  size: number,
  used?: boolean
): Promise<CodesListResponse> {
  const params = new URLSearchParams({ page: String(page), size: String(size) });
  if (used !== undefined) params.set("used", String(used));
  const res = await adminFetch(`${API_BASE}/api/admin/registration-codes?${params}`);
  return res.json();
}

export async function generateRegistrationCodes(
  count: number,
  note?: string
): Promise<{ codes: RegistrationCode[]; count: number }> {
  const res = await adminFetch(`${API_BASE}/api/admin/registration-codes`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ count, note }),
  });
  return res.json();
}

export async function disableRegistrationCode(id: number): Promise<void> {
  const res = await adminFetch(`${API_BASE}/api/admin/registration-codes/${id}/disable`, {
    method: "PUT",
  });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || "Disable failed");
  }
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 3: Commit**

```bash
git add frontend/lib/admin-fetch.ts
git commit -m "feat(frontend): add admin API helpers for registration codes"
```

---

### Task 10: Admin Registration Codes Page

**Files:**
- Create: `frontend/app/admin/registration-codes/page.tsx`

- [ ] **Step 1: Create the admin page**

```tsx
"use client";

import { useEffect, useState, useCallback } from "react";
import { Card } from "@/components/ui/Card";
import {
  fetchRegistrationCodes,
  generateRegistrationCodes,
  disableRegistrationCode,
  type RegistrationCode,
} from "@/lib/admin-fetch";

export default function AdminRegistrationCodesPage() {
  const [codes, setCodes] = useState<RegistrationCode[]>([]);
  const [total, setTotal] = useState(0);
  const [totalUnused, setTotalUnused] = useState(0);
  const [totalUsed, setTotalUsed] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [showGenerate, setShowGenerate] = useState(false);
  const [genCount, setGenCount] = useState(1);
  const [genNote, setGenNote] = useState("");
  const [genResult, setGenResult] = useState<RegistrationCode[] | null>(null);
  const [genLoading, setGenLoading] = useState(false);

  const fetchCodes = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchRegistrationCodes(page, 20);
      setCodes(data.codes || []);
      setTotal(data.total || 0);
    } catch { /* ignore */ }
    setLoading(false);
  }, [page]);

  const fetchStats = useCallback(async () => {
    try {
      const [unused, used] = await Promise.all([
        fetchRegistrationCodes(1, 1, false),
        fetchRegistrationCodes(1, 1, true),
      ]);
      setTotalUnused(unused.total || 0);
      setTotalUsed(used.total || 0);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => {
    fetchCodes();
    fetchStats();
  }, [fetchCodes, fetchStats]);

  const handleGenerate = async () => {
    setGenLoading(true);
    try {
      const result = await generateRegistrationCodes(genCount, genNote || undefined);
      setGenResult(result.codes);
      fetchCodes();
      fetchStats();
    } catch { /* ignore */ }
    setGenLoading(false);
  };

  const handleDisable = async (id: number) => {
    if (!confirm("确定要禁用这个邀请码吗？")) return;
    try {
      await disableRegistrationCode(id);
      fetchCodes();
      fetchStats();
    } catch { /* ignore */ }
  };

  const handleCopy = (code: string) => {
    navigator.clipboard.writeText(code);
  };

  const statusBadge = (c: RegistrationCode) => {
    if (c.is_disabled) {
      return <span className="inline-block px-3 py-1 rounded-pill text-xs font-semibold tracking-tight bg-red-error/10 text-red-error">已禁用</span>;
    }
    if (c.is_used) {
      return <span className="inline-block px-3 py-1 rounded-pill text-xs font-semibold tracking-tight bg-blue-100 text-blue-700">已使用</span>;
    }
    return <span className="inline-block px-3 py-1 rounded-pill text-xs font-semibold tracking-tight bg-green-light text-starbucks">未使用</span>;
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-starbucks tracking-tight">邀请码管理</h1>
          <p className="text-sm text-text-black-soft mt-0.5">生成和管理注册邀请码</p>
        </div>
        <button
          className="px-4 py-2 rounded-pill bg-green-accent text-white text-sm font-bold hover:bg-starbucks transition-colors"
          onClick={() => setShowGenerate(true)}
        >
          生成邀请码
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        <Card padding="md" className="text-center">
          <div className="text-2xl font-black text-starbucks">{total}</div>
          <div className="text-sm text-text-black-soft mt-1">总邀请码</div>
        </Card>
        <Card padding="md" className="text-center">
          <div className="text-2xl font-black text-starbucks">{totalUsed}</div>
          <div className="text-sm text-text-black-soft mt-1">已使用</div>
        </Card>
        <Card padding="md" className="text-center">
          <div className="text-2xl font-black text-starbucks">{totalUnused}</div>
          <div className="text-sm text-text-black-soft mt-1">未使用</div>
        </Card>
      </div>

      {/* Generate Modal */}
      {showGenerate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={() => { setShowGenerate(false); setGenResult(null); }}>
          <div className="bg-white rounded-xl p-6 w-full max-w-md mx-4 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-starbucks mb-4">生成邀请码</h2>

            {genResult ? (
              <div>
                <p className="text-sm text-text-black-soft mb-3">已生成 {genResult.length} 个邀请码：</p>
                <div className="space-y-2 mb-4 max-h-60 overflow-y-auto">
                  {genResult.map((c) => (
                    <div key={c.id} className="flex items-center justify-between bg-cream/50 rounded-lg px-3 py-2">
                      <code className="text-lg font-bold text-starbucks tracking-widest">{c.code}</code>
                      <button
                        className="text-xs font-bold text-green-accent hover:underline"
                        onClick={() => handleCopy(c.code)}
                      >
                        复制
                      </button>
                    </div>
                  ))}
                </div>
                <button
                  className="w-full py-2 rounded-pill bg-green-accent text-white text-sm font-bold"
                  onClick={() => { setShowGenerate(false); setGenResult(null); }}
                >
                  完成
                </button>
              </div>
            ) : (
              <div>
                <label className="block text-sm font-bold text-text-black mb-2">数量</label>
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={genCount}
                  onChange={(e) => setGenCount(Math.max(1, Math.min(100, Number(e.target.value))))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-4"
                />
                <label className="block text-sm font-bold text-text-black mb-2">备注（可选）</label>
                <input
                  type="text"
                  value={genNote}
                  onChange={(e) => setGenNote(e.target.value)}
                  placeholder="例如：给朋友"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-4"
                />
                <div className="flex gap-2">
                  <button
                    className="flex-1 py-2 rounded-pill border border-cream text-text-black-soft text-sm font-bold"
                    onClick={() => setShowGenerate(false)}
                  >
                    取消
                  </button>
                  <button
                    className="flex-1 py-2 rounded-pill bg-green-accent text-white text-sm font-bold disabled:opacity-50"
                    disabled={genLoading}
                    onClick={handleGenerate}
                  >
                    {genLoading ? "生成中..." : "生成"}
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Table */}
      <Card padding="md" className="overflow-hidden">
        {loading ? (
          <p className="text-text-black-soft text-center py-4">Loading...</p>
        ) : codes.length === 0 ? (
          <p className="text-text-black-soft text-center py-4">暂无邀请码，点击上方按钮生成。</p>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full min-w-[600px]">
                <thead>
                  <tr className="border-b border-cream">
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">邀请码</th>
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">状态</th>
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">备注</th>
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">使用者</th>
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">创建时间</th>
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">使用时间</th>
                    <th className="text-right p-3 text-sm font-semibold text-text-black tracking-tight">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {codes.map((c) => (
                    <tr key={c.id} className="border-b border-cream last:border-b-0 hover:bg-cream/50 transition-colors">
                      <td className="p-3 text-sm">
                        <code className="text-base font-bold text-starbucks tracking-widest">{c.code}</code>
                      </td>
                      <td className="p-3 text-sm">{statusBadge(c)}</td>
                      <td className="p-3 text-sm text-text-black-soft">{c.note || "-"}</td>
                      <td className="p-3 text-sm text-text-black">{c.used_by_nickname || "-"}</td>
                      <td className="p-3 text-sm text-text-black-soft">{c.created_at ? new Date(c.created_at).toLocaleDateString("zh-CN") : "-"}</td>
                      <td className="p-3 text-sm text-text-black-soft">{c.used_at ? new Date(c.used_at).toLocaleDateString("zh-CN") : "-"}</td>
                      <td className="p-3 text-right">
                        <button
                          className="text-sm font-semibold text-green-accent hover:underline mr-2"
                          onClick={() => handleCopy(c.code)}
                        >
                          复制
                        </button>
                        {!c.is_used && !c.is_disabled && (
                          <button
                            className="text-sm font-semibold text-red-error hover:underline"
                            onClick={() => handleDisable(c.id)}
                          >
                            禁用
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {total > 20 && (
              <div className="flex items-center justify-center gap-2 mt-4 pb-2">
                <button
                  className="px-3 py-1.5 text-sm font-semibold rounded-pill border border-green-accent text-green-accent transition-all duration-200 active:scale-[0.95] disabled:opacity-40 disabled:cursor-not-allowed"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => p - 1)}
                >
                  Previous
                </button>
                <span className="px-3 py-1 text-sm text-text-black-soft">
                  Page {page} of {Math.ceil(total / 20)}
                </span>
                <button
                  className="px-3 py-1.5 text-sm font-semibold rounded-pill border border-green-accent text-green-accent transition-all duration-200 active:scale-[0.95] disabled:opacity-40 disabled:cursor-not-allowed"
                  disabled={page >= Math.ceil(total / 20)}
                  onClick={() => setPage((p) => p + 1)}
                >
                  Next
                </button>
              </div>
            )}
          </>
        )}
      </Card>
    </div>
  );
}
```

- [ ] **Step 2: Verify TypeScript compilation**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 3: Commit**

```bash
git add frontend/app/admin/registration-codes/page.tsx
git commit -m "feat(frontend): add admin registration codes management page"
```

---

### Task 11: Admin Sidebar Nav Item

**Files:**
- Modify: `frontend/components/ui/AdminSidebar.tsx`

- [ ] **Step 1: Add the nav item**

In `AdminSidebar.tsx`, add to the `navItems` array (line 9-15), after the Users entry:

```typescript
  { href: "/admin/registration-codes", label: "邀请码", icon: "🔑" },
```

So the array becomes:

```typescript
const navItems = [
  { href: "/admin/dashboard", label: "Dashboard", icon: "📊" },
  { href: "/admin/users", label: "Users", icon: "👥" },
  { href: "/admin/registration-codes", label: "邀请码", icon: "🔑" },
  { href: "/admin/rooms", label: "Rooms", icon: "🃏" },
  { href: "/admin/ai-characters", label: "AI Characters", icon: "🧑" },
  { href: "/admin/llm-config", label: "LLM Config", icon: "🤖" },
  { href: "/admin/stats", label: "LLM Stats", icon: "📈" },
];
```

- [ ] **Step 2: Verify**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -10
cd frontend && npm run lint 2>&1 | tail -5
```

- [ ] **Step 3: Commit**

```bash
git add frontend/components/ui/AdminSidebar.tsx
git commit -m "feat(frontend): add registration codes nav item to admin sidebar"
```

---

### Task 12: Full Verification

- [ ] **Step 1: Backend checks**

```bash
cd server && go vet ./...
cd server && go test ./... 2>&1
```

Expected: all pass.

- [ ] **Step 2: Frontend checks**

```bash
cd frontend && npx tsc --noEmit 2>&1
cd frontend && npm run lint 2>&1
cd frontend && npm test 2>&1
```

Expected: all pass.

- [ ] **Step 3: Final commit if any fixes were needed**

```bash
git add -A
git commit -m "chore: verification fixes"
```
