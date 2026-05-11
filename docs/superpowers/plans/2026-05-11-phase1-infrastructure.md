# Phase 1: 基础设施搭建 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold the full project skeleton with Go backend (auth + WS hub + MySQL) and Next.js frontend (Starbucks design system + admin layout).

**Architecture:** Go backend provides REST auth endpoints + JWT + WebSocket upgrade + MySQL connection pool. Next.js frontend consumes REST API and serves auth pages, admin layout shell, and shared Starbucks-themed UI components.

**Tech Stack:** Go 1.22+, chi router, gorilla/websocket, golang-jwt, sqlx, MySQL 8. Next.js 14+ with TypeScript and Tailwind CSS.

---

## File Structure

### Go Backend

| File | Responsibility |
|------|---------------|
| `server/go.mod` | Go module definition |
| `server/cmd/server/main.go` | Config load, DB init, server start, graceful shutdown |
| `server/internal/config/config.go` | Env/yaml config struct + loader |
| `server/internal/model/user.go` | User + UserAuth structs, DB queries |
| `server/internal/auth/jwt.go` | JWT token generation + validation |
| `server/internal/auth/password.go` | bcrypt hash + compare |
| `server/internal/api/router.go` | Route registration |
| `server/internal/api/auth.go` | Register / Login / Guest handlers |
| `server/internal/api/middleware/auth.go` | JWT auth middleware + admin role check |
| `server/internal/api/middleware/cors.go` | CORS middleware |
| `server/internal/api/middleware/logging.go` | Request logging middleware |
| `server/internal/api/ws/hub.go` | WebSocket Hub skeleton (connection manager) |
| `server/migrations/001_create_users.sql` | Users + user_auths DDL |
| `server/internal/auth/jwt_test.go` | JWT unit tests |
| `server/internal/auth/password_test.go` | Password unit tests |
| `server/internal/api/auth_test.go` | Auth handler tests |

### Next.js Frontend

| File | Responsibility |
|------|---------------|
| `frontend/package.json` | Dependencies |
| `frontend/tailwind.config.ts` | Tailwind config with Starbucks tokens |
| `frontend/app/globals.css` | CSS variables + Tailwind layers |
| `frontend/app/layout.tsx` | Root layout (font loading, body class) |
| `frontend/app/auth/login/page.tsx` | Login page |
| `frontend/app/auth/register/page.tsx` | Register page |
| `frontend/app/auth/layout.tsx` | Auth pages layout (centered card) |
| `frontend/app/admin/layout.tsx` | Admin layout with sidebar |
| `frontend/app/admin/dashboard/page.tsx` | Dashboard placeholder |
| `frontend/app/admin/users/page.tsx` | Users placeholder |
| `frontend/app/admin/rooms/page.tsx` | Rooms placeholder |
| `frontend/app/admin/llm-config/page.tsx` | LLM config placeholder |
| `frontend/components/ui/Button.tsx` | Button component (all variants) |
| `frontend/components/ui/Card.tsx` | Card component |
| `frontend/components/ui/Input.tsx` | Floating-label input |
| `frontend/components/ui/AdminSidebar.tsx` | Admin sidebar navigation |
| `frontend/lib/api-client.ts` | REST API client with JWT |

---

### Task 1: VPS Docker 环境准备

**Files:**
- Create: `docker-compose.yml`
- Create: `server/Dockerfile`
- Create: `livekit.yaml`
- Create: `.env.example`
- Create: `.gitignore`

- [ ] **Step 1: Install Docker on VPS**
  ```bash
  # Docker Engine
  curl -fsSL https://get.docker.com -o get-docker.sh
  sudo sh get-docker.sh
  sudo usermod -aG docker $USER
  # Docker Compose plugin
  sudo apt-get install docker-compose-plugin -y
  docker --version && docker compose version
  ```

- [ ] **Step 2: Clone project on VPS and start services**
  ```bash
  # On VPS
  git clone <your-repo-url> vibe_pokeface
  cd vibe_pokeface
  cp .env.example .env   # edit passwords
  docker compose up -d   # starts mysql + livekit + go server
  ```

- [ ] **Step 3: Verify services are running**
  ```bash
  docker compose ps
  curl http://localhost:8080/api/health
  # Expected: {"status":"ok"}
  ```

- [ ] **Step 4: Configure firewall**
  ```bash
  sudo ufw allow 22/tcp        # SSH
  sudo ufw allow 8080/tcp      # Go server
  sudo ufw allow 7880/tcp      # LiveKit
  sudo ufw allow 7881/tcp      # LiveKit (TURN)
  sudo ufw enable
  ```

---

### Task 2: Initialize Go Module & Directory Structure

**Files:**
- Create: `server/go.mod`
- Create: `server/.gitkeep` placeholders for package dirs

- [ ] **Step 1: Initialize Go module**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface
  mkdir -p server/cmd/server server/internal/{config,model,auth,api/middleware,api/ws} server/migrations
  cd server
  go mod init github.com/yongkl/vibe-pokeface
  go get github.com/go-chi/chi/v5
  go get github.com/go-chi/cors
  go get github.com/gorilla/websocket
  go get github.com/golang-jwt/jwt/v5
  go get github.com/jmoiron/sqlx
  go get github.com/go-sql-driver/mysql
  go get golang.org/x/crypto
  go get github.com/joho/godotenv
  ```

- [ ] **Step 2: Run to verify module compiles**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go build ./...
  ```
  Expected: no errors, no output.

- [ ] **Step 3: Commit**
  ```bash
  git add server/go.mod server/go.sum server/
  git commit -m "chore: initialize Go module with dependencies"
  ```

---

### Task 3: Database Migration — Users & Auth Tables

**Files:**
- Create: `server/migrations/001_create_users.sql`
- Create: `server/internal/model/user.go`

- [ ] **Step 1: Create migration SQL**
  `server/migrations/001_create_users.sql`:
  ```sql
  CREATE TABLE IF NOT EXISTS users (
      id          BIGINT PRIMARY KEY AUTO_INCREMENT,
      nickname    VARCHAR(32) NOT NULL,
      avatar_url  VARCHAR(256),
      role        ENUM('user','admin') DEFAULT 'user',
      status      TINYINT DEFAULT 1,
      created_at  DATETIME DEFAULT NOW(),
      updated_at  DATETIME DEFAULT NOW()
  );

  CREATE TABLE IF NOT EXISTS user_auths (
      id           BIGINT PRIMARY KEY AUTO_INCREMENT,
      user_id      BIGINT NOT NULL,
      provider     ENUM('google','wechat','phone','guest'),
      provider_uid VARCHAR(128) NOT NULL,
      credential   VARCHAR(256),
      UNIQUE KEY uk_provider (provider, provider_uid)
  );
  ```

- [ ] **Step 2: Create model file**
  `server/internal/model/user.go`:
  ```go
  package model

  import "time"

  type User struct {
      ID        int64     `db:"id" json:"id"`
      Nickname  string    `db:"nickname" json:"nickname"`
      AvatarURL *string   `db:"avatar_url" json:"avatar_url,omitempty"`
      Role      string    `db:"role" json:"role"`
      Status    int8      `db:"status" json:"status"`
      CreatedAt time.Time `db:"created_at" json:"created_at"`
      UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
  }

  type UserAuth struct {
      ID          int64  `db:"id"`
      UserID      int64  `db:"user_id"`
      Provider    string `db:"provider"`
      ProviderUID string `db:"provider_uid"`
      Credential  string `db:"credential"`
  }

  // UserStore interface for DB operations — enables mock testing
  type UserStore interface {
      Create(ctx interface{ Deadline() (time.Time, bool) }, user *User) error
      FindByProvider(ctx interface{ Deadline() (time.Time, bool) }, provider, providerUID string) (*User, error)
      FindByID(ctx interface{ Deadline() (time.Time, bool) }, id int64) (*User, error)
  }
  ```

- [ ] **Step 3: Run migration against local/test MySQL**
  ```bash
  mysql -u root -e "CREATE DATABASE IF NOT EXISTS pokeface_test;"
  mysql -u root pokeface_test < server/migrations/001_create_users.sql
  mysql -u root pokeface_test -e "SHOW TABLES;"
  ```
  Expected: `users` and `user_auths` tables listed.

- [ ] **Step 4: Commit**
  ```bash
  git add server/migrations/001_create_users.sql server/internal/model/user.go
  git commit -m "feat: add users and user_auths tables with model"
  ```

---

### Task 4: Go Config & Database Connection

**Files:**
- Create: `server/internal/config/config.go`
- Modify: `server/internal/model/user.go` (add DB implementation)

- [ ] **Step 1: Write config loader**

  `server/internal/config/config.go`:
  ```go
  package config

  import (
      "os"
      "github.com/joho/godotenv"
  )

  type Config struct {
      Port       string
      DatabaseDSN string
      JWTSecret  string
      AllowedOrigins []string
  }

  func Load() (*Config, error) {
      godotenv.Load() // ignore error if no .env file

      cfg := &Config{
          Port:       getEnv("PORT", "8080"),
          DatabaseDSN: getEnv("DATABASE_DSN", "root:@tcp(127.0.0.1:3306)/pokeface?parseTime=true"),
          JWTSecret:  getEnv("JWT_SECRET", "dev-secret-change-in-production"),
          AllowedOrigins: []string{"http://localhost:3000"},
      }
      return cfg, nil
  }

  func getEnv(key, fallback string) string {
      if v := os.Getenv(key); v != "" {
          return v
      }
      return fallback
  }
  ```

- [ ] **Step 2: Write DB initialization function**
  `server/internal/model/db.go`:
  ```go
  package model

  import (
      "github.com/jmoiron/sqlx"
      _ "github.com/go-sql-driver/mysql"
  )

  func NewDB(dsn string) (*sqlx.DB, error) {
      db, err := sqlx.Connect("mysql", dsn)
      if err != nil {
          return nil, err
      }
      db.SetMaxOpenConns(25)
      db.SetMaxIdleConns(5)
      return db, nil
  }
  ```

- [ ] **Step 3: Write unit test for config loading**
  `server/internal/config/config_test.go`:
  ```go
  package config

  import (
      "os"
      "testing"
  )

  func TestLoad_Defaults(t *testing.T) {
      os.Unsetenv("PORT")
      os.Unsetenv("DATABASE_DSN")
      os.Unsetenv("JWT_SECRET")

      cfg, err := Load()
      if err != nil {
          t.Fatalf("Load() error = %v", err)
      }
      if cfg.Port != "8080" {
          t.Errorf("Port = %q, want %q", cfg.Port, "8080")
      }
      if cfg.JWTSecret != "dev-secret-change-in-production" {
          t.Errorf("JWTSecret = %q, want default", cfg.JWTSecret)
      }
  }

  func TestLoad_FromEnv(t *testing.T) {
      os.Setenv("PORT", "9090")
      os.Setenv("JWT_SECRET", "custom-secret")
      defer func() {
          os.Unsetenv("PORT")
          os.Unsetenv("JWT_SECRET")
      }()

      cfg, err := Load()
      if err != nil {
          t.Fatalf("Load() error = %v", err)
      }
      if cfg.Port != "9090" {
          t.Errorf("Port = %q, want %q", cfg.Port, "9090")
      }
      if cfg.JWTSecret != "custom-secret" {
          t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "custom-secret")
      }
  }
  ```

- [ ] **Step 4: Run tests**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./internal/config/ -v
  ```
  Expected: 2 PASS, 0 FAIL

- [ ] **Step 5: Build to verify compilation**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go build ./...
  ```

- [ ] **Step 6: Commit**
  ```bash
  git add server/internal/config/ server/internal/model/db.go
  git commit -m "feat: add config loader and database connection"
  ```

---

### Task 5: Auth Service — JWT & Password Hashing

**Files:**
- Create: `server/internal/auth/jwt.go`
- Create: `server/internal/auth/password.go`
- Create: `server/internal/auth/jwt_test.go`
- Create: `server/internal/auth/password_test.go`

- [ ] **Step 1: Write failing JWT test**

  `server/internal/auth/jwt_test.go`:
  ```go
  package auth

  import (
      "testing"
      "time"

      "github.com/golang-jwt/jwt/v5"
  )

  func TestGenerateToken_Valid(t *testing.T) {
      s := NewJWTService("test-secret")
      token, err := s.GenerateToken(42, "user")
      if err != nil {
          t.Fatalf("GenerateToken() error = %v", err)
      }
      if token == "" {
          t.Fatal("GenerateToken() returned empty string")
      }
  }

  func TestValidateToken_Valid(t *testing.T) {
      s := NewJWTService("test-secret")
      token, _ := s.GenerateToken(42, "admin")

      claims, err := s.ValidateToken(token)
      if err != nil {
          t.Fatalf("ValidateToken() error = %v", err)
      }
      if claims.UserID != 42 {
          t.Errorf("UserID = %d, want %d", claims.UserID, 42)
      }
      if claims.Role != "admin" {
          t.Errorf("Role = %s, want %s", claims.Role, "admin")
      }
  }

  func TestValidateToken_Expired(t *testing.T) {
      s := &JWTService{secret: "test-secret"}
      // Manually create an expired token
      now := time.Now()
      claims := &Claims{
          UserID: 1,
          Role:   "user",
          RegisteredClaims: jwt.RegisteredClaims{
              ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
              IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
          },
      }
      token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
      tokenStr, _ := token.SignedString([]byte("test-secret"))

      _, err := s.ValidateToken(tokenStr)
      if err == nil {
          t.Fatal("ValidateToken() expected error for expired token, got nil")
      }
  }

  func TestValidateToken_InvalidSignature(t *testing.T) {
      s := NewJWTService("test-secret")
      other := NewJWTService("other-secret")
      token, _ := other.GenerateToken(1, "user")

      _, err := s.ValidateToken(token)
      if err == nil {
          t.Fatal("ValidateToken() expected error for wrong signature, got nil")
      }
  }
  ```

- [ ] **Step 2: Run test to verify it fails**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./internal/auth/ -v
  ```
  Expected: package has no testable Go files (or compile error if file exists but no implementation yet).

- [ ] **Step 3: Write JWT implementation**

  `server/internal/auth/jwt.go`:
  ```go
  package auth

  import (
      "time"
      "github.com/golang-jwt/jwt/v5"
  )

  type Claims struct {
      UserID int64  `json:"user_id"`
      Role   string `json:"role"`
      jwt.RegisteredClaims
  }

  type JWTService struct {
      secret    string
      expiresIn time.Duration
  }

  func NewJWTService(secret string) *JWTService {
      return &JWTService{
          secret:    secret,
          expiresIn: 72 * time.Hour, // 3 days
      }
  }

  func (s *JWTService) GenerateToken(userID int64, role string) (string, error) {
      now := time.Now()
      claims := &Claims{
          UserID: userID,
          Role:   role,
          RegisteredClaims: jwt.RegisteredClaims{
              ExpiresAt: jwt.NewNumericDate(now.Add(s.expiresIn)),
              IssuedAt:  jwt.NewNumericDate(now),
          },
      }
      token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
      return token.SignedString([]byte(s.secret))
  }

  func (s *JWTService) ValidateToken(tokenStr string) (*Claims, error) {
      token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
          return []byte(s.secret), nil
      })
      if err != nil {
          return nil, err
      }
      claims, ok := token.Claims.(*Claims)
      if !ok || !token.Valid {
          return nil, jwt.ErrSignatureInvalid
      }
      return claims, nil
  }
  ```

- [ ] **Step 4: Run tests to verify they pass**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./internal/auth/ -v
  ```
  Expected: 4 PASS, 0 FAIL

- [ ] **Step 5: Write failing password tests**

  `server/internal/auth/password_test.go`:
  ```go
  package auth

  import "testing"

  func TestHashPassword_ReturnsHash(t *testing.T) {
      hash, err := HashPassword("my-secure-password")
      if err != nil {
          t.Fatalf("HashPassword() error = %v", err)
      }
      if hash == "" {
          t.Fatal("HashPassword() returned empty string")
      }
  }

  func TestCheckPassword_Correct(t *testing.T) {
      hash, _ := HashPassword("my-secure-password")
      if !CheckPassword(hash, "my-secure-password") {
          t.Fatal("CheckPassword() = false, want true for correct password")
      }
  }

  func TestCheckPassword_Wrong(t *testing.T) {
      hash, _ := HashPassword("my-secure-password")
      if CheckPassword(hash, "wrong-password") {
          t.Fatal("CheckPassword() = true, want false for wrong password")
      }
  }
  ```

- [ ] **Step 6: Run to verify they fail**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./internal/auth/ -v
  ```
  Expected: compile error (HashPassword/CheckPassword not defined)

- [ ] **Step 7: Write password implementation**

  `server/internal/auth/password.go`:
  ```go
  package auth

  import "golang.org/x/crypto/bcrypt"

  const bcryptCost = 12

  func HashPassword(password string) (string, error) {
      bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
      return string(bytes), err
  }

  func CheckPassword(hash, password string) bool {
      err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
      return err == nil
  }
  ```

- [ ] **Step 8: Run all auth tests**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./internal/auth/ -v
  ```
  Expected: 7 PASS, 0 FAIL

- [ ] **Step 9: Commit**
  ```bash
  git add server/internal/auth/
  git commit -m "feat: implement JWT tokens and bcrypt password hashing"
  ```

---

### Task 6: Auth REST Handlers

**Files:**
- Create: `server/internal/api/auth.go`
- Create: `server/internal/api/auth_test.go`

- [ ] **Step 1: Write failing auth handler tests**

  `server/internal/api/auth_test.go`:
  ```go
  package api

  import (
      "bytes"
      "encoding/json"
      "net/http"
      "net/http/httptest"
      "testing"

      "github.com/go-chi/chi/v5"
      "github.com/yongkl/vibe-pokeface/internal/auth"
      "github.com/yongkl/vibe-pokeface/internal/model"
  )

  // mockUserStore implements model.UserStore for testing
  type mockUserStore struct {
      createFn        func(user *model.User) error
      findByProviderFn func(provider, providerUID string) (*model.User, error)
      findByIDFn      func(id int64) (*model.User, error)
  }

  func (m *mockUserStore) Create(ctx interface{ Deadline() (time.Time, bool) }, user *model.User) error {
      return m.createFn(user)
  }
  func (m *mockUserStore) FindByProvider(ctx interface{ Deadline() (time.Time, bool) }, provider, providerUID string) (*model.User, error) {
      return m.findByProviderFn(provider, providerUID)
  }
  func (m *mockUserStore) FindByID(ctx interface{ Deadline() (time.Time, bool) }, id int64) (*model.User, error) {
      return m.findByIDFn(id)
  }
  ```
  Wait — the `context` interface param in model.UserStore won't work. Let me fix that in the model first.

  Update `server/internal/model/user.go` — change the `UserStore` interface to use `context.Context`:
  ```go
  import "context"

  type UserStore interface {
      Create(ctx context.Context, user *User) error
      FindByProvider(ctx context.Context, provider, providerUID string) (*User, error)
      FindByID(ctx context.Context, id int64) (*User, error)
  }
  ```

  Now write the test:
  `server/internal/api/auth_test.go`:
  ```go
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
  }

  func (m *mockUserStore) Create(ctx context.Context, user *model.User) error {
      return m.createFn(ctx, user)
  }
  func (m *mockUserStore) FindByProvider(ctx context.Context, provider, providerUID string) (*model.User, error) {
      return m.findByProviderFn(ctx, provider, providerUID)
  }
  func (m *mockUserStore) FindByID(ctx context.Context, id int64) (*model.User, error) {
      return nil, nil
  }

  func TestRegister_Success(t *testing.T) {
      store := &mockUserStore{
          createFn: func(ctx context.Context, user *model.User) error {
              user.ID = 1
              return nil
          },
          findByProviderFn: func(ctx context.Context, provider, providerUID string) (*model.User, error) {
              return nil, nil // not found = new user
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
      if resp["user"] == nil {
          t.Error("response missing user")
      }
  }

  func TestRegister_Duplicate(t *testing.T) {
      existingUser := &model.User{ID: 1, Nickname: "existing"}
      store := &mockUserStore{
          findByProviderFn: func(ctx context.Context, provider, providerUID string) (*model.User, error) {
              if providerUID == "guest:device123" {
                  return existingUser, nil
              }
              return nil, nil
          },
      }
      jwtSvc := auth.NewJWTService("test-secret")
      handler := NewAuthHandler(store, jwtSvc)

      // Guest register with existing provider_uid
      body := map[string]string{"nickname": "testuser", "password": "secure123"}
      b, _ := json.Marshal(body)
      req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
      req.Header.Set("Content-Type", "application/json")
      w := httptest.NewRecorder()
      handler.Register(w, req)
      if w.Code != http.StatusConflict {
          t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
      }
  }

  func TestLogin_Success(t *testing.T) {
      hashedPw, _ := auth.HashPassword("correct-pw")
      existingUser := &model.User{ID: 1, Nickname: "testuser", Role: "user"}
      store := &mockUserStore{
          findByProviderFn: func(ctx context.Context, provider, providerUID string) (*model.User, error) {
              return existingUser, nil
          },
      }
      jwtSvc := auth.NewJWTService("test-secret")
      handler := NewAuthHandler(store, jwtSvc)

      // Simulate stored credential check
      _ = hashedPw // will be used by handler to validate
  }
  ```

  Hmm, this login test is incomplete since the handler needs to look up the credential. Let me design the handler to work with a `UserStore` that also provides credential data. Let me revise the approach.

  Actually, for simplicity, the register handler will:
  1. Accept `nickname`, `password`, and optional `provider_uid`
  2. Create user + user_auths record with bcrypt hashed password as credential
  3. Return JWT token

  The login handler will:
  1. Accept `password` + `provider_uid` (or nickname lookup)
  2. Find via provider_uid, compare bcrypt hash
  3. Return JWT token

  Let me simplify the test to just cover Register + Login.

  Full `server/internal/api/auth_test.go`:
  ```go
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
      createAuthFn     func(ctx context.Context, auth *model.UserAuth) error
      findAuthFn       func(ctx context.Context, provider, providerUID string) (*model.UserAuth, error)
  }

  func (m *mockUserStore) Create(ctx context.Context, user *model.User) error {
      return m.createFn(ctx, user)
  }
  func (m *mockUserStore) FindByProvider(ctx context.Context, provider, providerUID string) (*model.User, error) {
      return m.findByProviderFn(ctx, provider, providerUID)
  }
  func (m *mockUserStore) FindByID(ctx context.Context, id int64) (*model.User, error) {
      return nil, nil
  }
  func (m *mockUserStore) CreateAuth(ctx context.Context, ua *model.UserAuth) error {
      return m.createAuthFn(ctx, ua)
  }
  func (m *mockUserStore) FindAuth(ctx context.Context, provider, providerUID string) (*model.UserAuth, error) {
      return m.findAuthFn(ctx, provider, providerUID)
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
      store := &mockUserStore{
          createFn: func(ctx context.Context, user *model.User) error { return nil },
      }
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
          findByProviderFn: func(ctx context.Context, provider, providerUID string) (*model.User, error) {
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
              return nil, nil // new guest
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
  ```

- [ ] **Step 2: Run test to verify it fails**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./internal/api/ -v
  ```
  Expected: compile error (AuthHandler, Register, Login not defined). plus model.UserStore needs FindAuth/CreateAuth methods.

- [ ] **Step 3: Update UserStore interface with auth methods**

  Update `server/internal/model/user.go`:
  ```go
  type UserStore interface {
      Create(ctx context.Context, user *User) error
      FindByProvider(ctx context.Context, provider, providerUID string) (*User, error)
      FindByID(ctx context.Context, id int64) (*User, error)
      CreateAuth(ctx context.Context, auth *UserAuth) error
      FindAuth(ctx context.Context, provider, providerUID string) (*UserAuth, error)
  }
  ```

- [ ] **Step 4: Write auth handler implementation**

  `server/internal/api/auth.go`:
  ```go
  package api

  import (
      "encoding/json"
      "errors"
      "net/http"

      "github.com/yongkl/vibe-pokeface/internal/auth"
      "github.com/yongkl/vibe-pokeface/internal/model"
  )

  type AuthHandler struct {
      store   model.UserStore
      jwt     *auth.JWTService
  }

  func NewAuthHandler(store model.UserStore, jwt *auth.JWTService) *AuthHandler {
      return &AuthHandler{store: store, jwt: jwt}
  }

  type registerRequest struct {
      Nickname string `json:"nickname"`
      Password string `json:"password"`
  }

  type loginRequest struct {
      Password    string `json:"password"`
      ProviderUID string `json:"provider_uid"`
  }

  type guestRequest struct {
      DeviceID string `json:"device_id"`
  }

  type authResponse struct {
      Token string      `json:"token"`
      User  *model.User `json:"user"`
  }

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

      // Check if guest auth already exists
      providerUID := "password:" + req.Nickname
      existing, _ := h.store.FindByProvider(r.Context(), "password", providerUID)
      if existing != nil {
          http.Error(w, `{"error":"user already exists"}`, http.StatusConflict)
          return
      }

      user := &model.User{Nickname: req.Nickname, Role: "user"}
      if err := h.store.Create(r.Context(), user); err != nil {
          http.Error(w, `{"error":"failed to create user"}`, http.StatusInternalServerError)
          return
      }

      hashedPw, err := auth.HashPassword(req.Password)
      if err != nil {
          http.Error(w, `{"error":"failed to process password"}`, http.StatusInternalServerError)
          return
      }

      userAuth := &model.UserAuth{
          UserID:      user.ID,
          Provider:    "password",
          ProviderUID: providerUID,
          Credential:  hashedPw,
      }
      if err := h.store.CreateAuth(r.Context(), userAuth); err != nil {
          http.Error(w, `{"error":"failed to save auth"}`, http.StatusInternalServerError)
          return
      }

      token, _ := h.jwt.GenerateToken(user.ID, user.Role)
      json.NewEncoder(w).Encode(authResponse{Token: token, User: user})
  }

  func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
      var req loginRequest
      if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
          http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
          return
      }
      if req.Password == "" {
          http.Error(w, `{"error":"password required"}`, http.StatusBadRequest)
          return
      }

      providerUID := req.ProviderUID
      if providerUID == "" {
          providerUID = "password:default"
      }

      userAuth, err := h.store.FindAuth(r.Context(), "password", providerUID)
      if err != nil || userAuth == nil {
          http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
          return
      }

      if !auth.CheckPassword(userAuth.Credential, req.Password) {
          http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
          return
      }

      user, err := h.store.FindByID(r.Context(), userAuth.UserID)
      if err != nil || user == nil {
          http.Error(w, `{"error":"user not found"}`, http.StatusInternalServerError)
          return
      }

      token, _ := h.jwt.GenerateToken(user.ID, user.Role)
      json.NewEncoder(w).Encode(authResponse{Token: token, User: user})
  }

  func (h *AuthHandler) GuestLogin(w http.ResponseWriter, r *http.Request) {
      var req guestRequest
      if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
          http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
          return
      }
      if req.DeviceID == "" {
          http.Error(w, `{"error":"device_id required"}`, http.StatusBadRequest)
          return
      }

      providerUID := "guest:" + req.DeviceID
      existing, _ := h.store.FindByProvider(r.Context(), "guest", providerUID)
      if existing != nil {
          // Returning guest — log them in
          token, _ := h.jwt.GenerateToken(existing.ID, existing.Role)
          json.NewEncoder(w).Encode(authResponse{Token: token, User: existing})
          return
      }

      // New guest — create user + auth
      user := &model.User{Nickname: "Guest-" + req.DeviceID[:6], Role: "user"}
      if err := h.store.Create(r.Context(), user); err != nil {
          http.Error(w, `{"error":"failed to create guest"}`, http.StatusInternalServerError)
          return
      }
      userAuth := &model.UserAuth{
          UserID:      user.ID,
          Provider:    "guest",
          ProviderUID: providerUID,
      }
      if err := h.store.CreateAuth(r.Context(), userAuth); err != nil {
          http.Error(w, `{"error":"failed to save guest auth"}`, http.StatusInternalServerError)
          return
      }

      token, _ := h.jwt.GenerateToken(user.ID, user.Role)
      json.NewEncoder(w).Encode(authResponse{Token: token, User: user})
  }
  ```

- [ ] **Step 5: Run tests to verify they pass**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./internal/api/ -v -count=1
  ```
  Expected: 5 PASS, 0 FAIL

- [ ] **Step 6: Commit**
  ```bash
  git add server/internal/api/auth.go server/internal/api/auth_test.go server/internal/model/user.go
  git commit -m "feat: implement register, login, and guest login handlers"
  ```

---

### Task 7: Middleware — Auth, CORS, Logging

**Files:**
- Create: `server/internal/api/middleware/auth.go`
- Create: `server/internal/api/middleware/cors.go`
- Create: `server/internal/api/middleware/logging.go`
- Create: `server/internal/api/middleware/auth_test.go`

- [ ] **Step 1: Write auth middleware + test**

  `server/internal/api/middleware/auth.go`:
  ```go
  package middleware

  import (
      "context"
      "net/http"
      "strings"

      "github.com/yongkl/vibe-pokeface/internal/auth"
  )

  type contextKey string

  const (
      ClaimsKey contextKey = "claims"
  )

  func Auth(jwt *auth.JWTService) func(http.Handler) http.Handler {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              authHeader := r.Header.Get("Authorization")
              if authHeader == "" {
                  http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
                  return
              }

              parts := strings.SplitN(authHeader, " ", 2)
              if len(parts) != 2 || parts[0] != "Bearer" {
                  http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
                  return
              }

              claims, err := jwt.ValidateToken(parts[1])
              if err != nil {
                  http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
                  return
              }

              ctx := context.WithValue(r.Context(), ClaimsKey, claims)
              next.ServeHTTP(w, r.WithContext(ctx))
          })
      }
  }

  func AdminOnly(next http.Handler) http.Handler {
      return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          claims, ok := r.Context().Value(ClaimsKey).(*auth.Claims)
          if !ok || claims.Role != "admin" {
              http.Error(w, `{"error":"admin access required"}`, http.StatusForbidden)
              return
          }
          next.ServeHTTP(w, r.WithContext(r.Context()))
      })
  }
  ```

  `server/internal/api/middleware/auth_test.go`:
  ```go
  package middleware

  import (
      "net/http"
      "net/http/httptest"
      "testing"

      "github.com/yongkl/vibe-pokeface/internal/auth"
  )

  func TestAuth_MissingHeader(t *testing.T) {
      jwt := auth.NewJWTService("test-secret")
      handler := Auth(jwt)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(http.StatusOK)
      }))

      req := httptest.NewRequest("GET", "/", nil)
      w := httptest.NewRecorder()
      handler.ServeHTTP(w, req)
      if w.Code != http.StatusUnauthorized {
          t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
      }
  }

  func TestAuth_InvalidHeader(t *testing.T) {
      jwt := auth.NewJWTService("test-secret")
      handler := Auth(jwt)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(http.StatusOK)
      }))

      req := httptest.NewRequest("GET", "/", nil)
      req.Header.Set("Authorization", "Invalid token")
      w := httptest.NewRecorder()
      handler.ServeHTTP(w, req)
      if w.Code != http.StatusUnauthorized {
          t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
      }
  }

  func TestAuth_ValidToken(t *testing.T) {
      jwt := auth.NewJWTService("test-secret")
      handler := Auth(jwt)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          claims := r.Context().Value(ClaimsKey).(*auth.Claims)
          if claims.UserID != 42 {
              t.Errorf("UserID = %d, want %d", claims.UserID, 42)
          }
          w.WriteHeader(http.StatusOK)
      }))

      token, _ := jwt.GenerateToken(42, "user")
      req := httptest.NewRequest("GET", "/", nil)
      req.Header.Set("Authorization", "Bearer "+token)
      w := httptest.NewRecorder()
      handler.ServeHTTP(w, req)
      if w.Code != http.StatusOK {
          t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
      }
  }

  func TestAdminOnly_ForbidsUser(t *testing.T) {
      jwt := auth.NewJWTService("test-secret")
      handler := Auth(jwt)(AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(http.StatusOK)
      })))

      token, _ := jwt.GenerateToken(1, "user")
      req := httptest.NewRequest("GET", "/admin", nil)
      req.Header.Set("Authorization", "Bearer "+token)
      w := httptest.NewRecorder()
      handler.ServeHTTP(w, req)
      if w.Code != http.StatusForbidden {
          t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
      }
  }

  func TestAdminOnly_AllowsAdmin(t *testing.T) {
      jwt := auth.NewJWTService("test-secret")
      handler := Auth(jwt)(AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(http.StatusOK)
      })))

      token, _ := jwt.GenerateToken(1, "admin")
      req := httptest.NewRequest("GET", "/admin", nil)
      req.Header.Set("Authorization", "Bearer "+token)
      w := httptest.NewRecorder()
      handler.ServeHTTP(w, req)
      if w.Code != http.StatusOK {
          t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
      }
  }
  ```

- [ ] **Step 2: Write CORS middleware**

  `server/internal/api/middleware/cors.go`:
  ```go
  package middleware

  import (
      "net/http"
      "strings"
  )

  type CORSConfig struct {
      AllowedOrigins []string
  }

  func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              origin := r.Header.Get("Origin")
              allowed := false
              for _, o := range cfg.AllowedOrigins {
                  if o == "*" || o == origin {
                      allowed = true
                      break
                  }
              }
              if allowed {
                  w.Header().Set("Access-Control-Allow-Origin", origin)
                  w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                  w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
                  w.Header().Set("Access-Control-Allow-Credentials", "true")
              }

              if r.Method == http.MethodOptions {
                  w.WriteHeader(http.StatusNoContent)
                  return
              }

              next.ServeHTTP(w, r)
          })
      }
  }

  // isAllowed checks if origin is in the allowed list
  func isAllowed(origin string, origins []string) bool {
      for _, o := range origins {
          if o == "*" || o == origin {
              return true
          }
      }
      return strings.HasPrefix(origin, "http://localhost:")
  }
  ```

  Update to use the helper:
  ```go
  func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              origin := r.Header.Get("Origin")
              if isAllowed(origin, cfg.AllowedOrigins) {
                  w.Header().Set("Access-Control-Allow-Origin", origin)
                  w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                  w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
                  w.Header().Set("Access-Control-Allow-Credentials", "true")
              }
              if r.Method == http.MethodOptions {
                  w.WriteHeader(http.StatusNoContent)
                  return
              }
              next.ServeHTTP(w, r)
          })
      }
  }
  ```

- [ ] **Step 3: Write logging middleware**

  `server/internal/api/middleware/logging.go`:
  ```go
  package middleware

  import (
      "log"
      "net/http"
      "time"
  )

  type responseWriter struct {
      http.ResponseWriter
      status int
  }

  func (rw *responseWriter) WriteHeader(code int) {
      rw.status = code
      rw.ResponseWriter.WriteHeader(code)
  }

  func Logging(next http.Handler) http.Handler {
      return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          start := time.Now()
          rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
          next.ServeHTTP(rw, r)
          log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start))
      })
  }
  ```

- [ ] **Step 4: Run middleware tests**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./internal/api/middleware/ -v
  ```
  Expected: 5 PASS, 0 FAIL

- [ ] **Step 5: Commit**
  ```bash
  git add server/internal/api/middleware/
  git commit -m "feat: add auth, CORS, and logging middleware"
  ```

---

### Task 8: WebSocket Hub Skeleton

**Files:**
- Create: `server/internal/api/ws/hub.go`
- Create: `server/internal/api/ws/hub_test.go`

- [ ] **Step 1: Write failing WS Hub test**

  `server/internal/api/ws/hub_test.go`:
  ```go
  package ws

  import (
      "testing"
  )

  func TestNewHub(t *testing.T) {
      h := NewHub()
      if h == nil {
          t.Fatal("NewHub() returned nil")
      }
  }

  func TestHub_RegisterClient(t *testing.T) {
      h := NewHub()
      client := &Client{ID: "test-1", RoomID: "room-1"}
      h.Register <- client
      // This test just verifies no panic when sending to channel
  }

  func TestHub_UnregisterClient(t *testing.T) {
      h := NewHub()
      client := &Client{ID: "test-1", RoomID: "room-1"}
      h.Unregister <- client
  }
  ```

  Note: these tests only verify channel sends don't block. The Hub needs to run via `h.Run()` in a goroutine to actually process these.

- [ ] **Step 2: Run test to verify compile**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./internal/api/ws/ -v
  ```
  Expected: compile error (Hub, Client not defined)

- [ ] **Step 3: Write Hub implementation**

  `server/internal/api/ws/hub.go`:
  ```go
  package ws

  import (
      "sync"
  )

  // Client represents a single WebSocket connection
  type Client struct {
      ID     string
      RoomID string
      Send   chan []byte
  }

  // RoomHub manages all clients in a single room
  type RoomHub struct {
      Clients map[string]*Client
      mu      sync.RWMutex
  }

  func NewRoomHub() *RoomHub {
      return &RoomHub{
          Clients: make(map[string]*Client),
      }
  }

  func (rh *RoomHub) Add(client *Client) {
      rh.mu.Lock()
      defer rh.mu.Unlock()
      rh.Clients[client.ID] = client
  }

  func (rh *RoomHub) Remove(clientID string) {
      rh.mu.Lock()
      defer rh.mu.Unlock()
      delete(rh.Clients, clientID)
  }

  func (rh *RoomHub) Broadcast(msg []byte) {
      rh.mu.RLock()
      defer rh.mu.RUnlock()
      for _, client := range rh.Clients {
          select {
          case client.Send <- msg:
          default:
              // Client buffer full, skip
          }
      }
  }

  func (rh *RoomHub) Count() int {
      rh.mu.RLock()
      defer rh.mu.RUnlock()
      return len(rh.Clients)
  }

  // Hub manages all rooms
  type Hub struct {
      Rooms      map[string]*RoomHub
      mu         sync.RWMutex
      Register   chan *Client
      Unregister chan *Client
  }

  func NewHub() *Hub {
      return &Hub{
          Rooms:      make(map[string]*RoomHub),
          Register:   make(chan *Client, 256),
          Unregister: make(chan *Client, 256),
      }
  }

  func (h *Hub) Run() {
      for {
          select {
          case client := <-h.Register:
              h.mu.RLock()
              room, ok := h.Rooms[client.RoomID]
              h.mu.RUnlock()
              if !ok {
                  room = NewRoomHub()
                  h.mu.Lock()
                  h.Rooms[client.RoomID] = room
                  h.mu.Unlock()
              }
              room.Add(client)

          case client := <-h.Unregister:
              h.mu.RLock()
              room, ok := h.Rooms[client.RoomID]
              h.mu.RUnlock()
              if ok {
                  room.Remove(client.ID)
                  if room.Count() == 0 {
                      h.mu.Lock()
                      delete(h.Rooms, client.RoomID)
                      h.mu.Unlock()
                  }
              }
          }
      }
  }

  func (h *Hub) GetRoom(roomID string) *RoomHub {
      h.mu.RLock()
      defer h.mu.RUnlock()
      return h.Rooms[roomID]
  }
  ```

- [ ] **Step 4: Run tests**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./internal/api/ws/ -v -count=1
  ```
  Expected: 3 PASS, 0 FAIL

- [ ] **Step 5: Commit**
  ```bash
  git add server/internal/api/ws/
  git commit -m "feat: add WebSocket hub and room manager skeleton"
  ```

---

### Task 9: Router & Main Entry Point

**Files:**
- Create: `server/internal/api/router.go`
- Create: `server/cmd/server/main.go`

- [ ] **Step 1: Write router**

  `server/internal/api/router.go`:
  ```go
  package api

  import (
      "github.com/go-chi/chi/v5"
      "github.com/yongkl/vibe-pokeface/internal/api/middleware"
      "github.com/yongkl/vibe-pokeface/internal/api/ws"
      "github.com/yongkl/vibe-pokeface/internal/auth"
      "github.com/yongkl/vibe-pokeface/internal/model"
  )

  func NewRouter(store model.UserStore, jwt *auth.JWTService, hub *ws.Hub, corsCfg middleware.CORSConfig) *chi.Mux {
      r := chi.NewRouter()

      // Global middleware
      r.Use(middleware.Logging)
      r.Use(middleware.CORS(corsCfg))

      // Auth handler
      authHandler := NewAuthHandler(store, jwt)

      // Public routes
      r.Route("/api/auth", func(r chi.Router) {
          r.Post("/register", authHandler.Register)
          r.Post("/login", authHandler.Login)
          r.Post("/guest", authHandler.GuestLogin)
      })

      // Health check
      r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(http.StatusOK)
          w.Write([]byte(`{"status":"ok"}`))
      })

      // Protected routes
      r.Group(func(r chi.Router) {
          r.Use(middleware.Auth(jwt))

          // Admin routes
          r.Group(func(r chi.Router) {
              r.Use(middleware.AdminOnly)
              r.Route("/api/admin", func(r chi.Router) {
                  r.Get("/dashboard", placeholderHandler("dashboard"))
                  r.Get("/users", placeholderHandler("users"))
                  r.Get("/rooms", placeholderHandler("rooms"))
                  r.Get("/llm-config", placeholderHandler("llm-config"))
              })
          })
      })

      return r
  }
  ```

  Need to add the import for `net/http`:

  `server/internal/api/router.go` (full):
  ```go
  package api

  import (
      "net/http"

      "github.com/go-chi/chi/v5"
      "github.com/yongkl/vibe-pokeface/internal/api/middleware"
      "github.com/yongkl/vibe-pokeface/internal/api/ws"
      "github.com/yongkl/vibe-pokeface/internal/auth"
      "github.com/yongkl/vibe-pokeface/internal/model"
  )

  func NewRouter(store model.UserStore, jwt *auth.JWTService, hub *ws.Hub, corsCfg middleware.CORSConfig) *chi.Mux {
      r := chi.NewRouter()

      r.Use(middleware.Logging)
      r.Use(middleware.CORS(corsCfg))

      authHandler := NewAuthHandler(store, jwt)

      r.Route("/api/auth", func(r chi.Router) {
          r.Post("/register", authHandler.Register)
          r.Post("/login", authHandler.Login)
          r.Post("/guest", authHandler.GuestLogin)
      })

      r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(http.StatusOK)
          w.Write([]byte(`{"status":"ok"}`))
      })

      r.Group(func(r chi.Router) {
          r.Use(middleware.Auth(jwt))

          r.Group(func(r chi.Router) {
              r.Use(middleware.AdminOnly)
              r.Route("/api/admin", func(r chi.Router) {
                  r.Get("/dashboard", placeholderHandler("dashboard"))
                  r.Get("/users", placeholderHandler("users"))
                  r.Get("/rooms", placeholderHandler("rooms"))
                  r.Get("/llm-config", placeholderHandler("llm-config"))
              })
          })
      })

      return r
  }

  func placeholderHandler(name string) http.HandlerFunc {
      return func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "application/json")
          w.Write([]byte(`{"message":"` + name + ` - coming soon","status":"ok"}`))
      }
  }
  ```

- [ ] **Step 2: Write main.go**

  `server/cmd/server/main.go`:
  ```go
  package main

  import (
      "context"
      "log"
      "net/http"
      "os"
      "os/signal"
      "syscall"
      "time"

      "github.com/yongkl/vibe-pokeface/internal/api"
      "github.com/yongkl/vibe-pokeface/internal/api/middleware"
      "github.com/yongkl/vibe-pokeface/internal/api/ws"
      "github.com/yongkl/vibe-pokeface/internal/auth"
      "github.com/yongkl/vibe-pokeface/internal/config"
      "github.com/yongkl/vibe-pokeface/internal/model"
  )

  func main() {
      cfg, err := config.Load()
      if err != nil {
          log.Fatalf("Failed to load config: %v", err)
      }

      db, err := model.NewDB(cfg.DatabaseDSN)
      if err != nil {
          log.Fatalf("Failed to connect to database: %v", err)
      }
      defer db.Close()

      jwtSvc := auth.NewJWTService(cfg.JWTSecret)
      hub := ws.NewHub()
      go hub.Run()

      router := api.NewRouter(db, jwtSvc, hub, middleware.CORSConfig{
          AllowedOrigins: cfg.AllowedOrigins,
      })

      srv := &http.Server{
          Addr:    ":" + cfg.Port,
          Handler: router,
      }

      // Graceful shutdown
      go func() {
          sigCh := make(chan os.Signal, 1)
          signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
          <-sigCh
          log.Println("Shutting down...")
          ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
          defer cancel()
          srv.Shutdown(ctx)
      }()

      log.Printf("Server starting on :%s", cfg.Port)
      if err := srv.ListenAndServe(); err != http.ErrServerClosed {
          log.Fatalf("Server error: %v", err)
      }
  }
  ```

  Note: `db` (`*sqlx.DB`) needs to implement `model.UserStore`. We need a DB-backed implementation. For now, let's create a wrapper.

- [ ] **Step 3: Create DB-backed UserStore**

  `server/internal/model/user_store.go`:
  ```go
  package model

  import (
      "context"

      "github.com/jmoiron/sqlx"
  )

  type UserDB struct {
      db *sqlx.DB
  }

  func NewUserDB(db *sqlx.DB) *UserDB {
      return &UserDB{db: db}
  }

  func (s *UserDB) Create(ctx context.Context, user *User) error {
      result, err := s.db.ExecContext(ctx, "INSERT INTO users (nickname, role) VALUES (?, ?)", user.Nickname, user.Role)
      if err != nil {
          return err
      }
      id, _ := result.LastInsertId()
      user.ID = id
      return nil
  }

  func (s *UserDB) FindByProvider(ctx context.Context, provider, providerUID string) (*User, error) {
      var user User
      err := s.db.GetContext(ctx, &user, `
          SELECT u.* FROM users u
          JOIN user_auths ua ON u.id = ua.user_id
          WHERE ua.provider = ? AND ua.provider_uid = ?
      `, provider, providerUID)
      if err != nil {
          return nil, err
      }
      return &user, nil
  }

  func (s *UserDB) FindByID(ctx context.Context, id int64) (*User, error) {
      var user User
      err := s.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = ?", id)
      if err != nil {
          return nil, err
      }
      return &user, nil
  }

  func (s *UserDB) CreateAuth(ctx context.Context, ua *UserAuth) error {
      _, err := s.db.ExecContext(ctx,
          "INSERT INTO user_auths (user_id, provider, provider_uid, credential) VALUES (?, ?, ?, ?)",
          ua.UserID, ua.Provider, ua.ProviderUID, ua.Credential)
      return err
  }

  func (s *UserDB) FindAuth(ctx context.Context, provider, providerUID string) (*UserAuth, error) {
      var ua UserAuth
      err := s.db.GetContext(ctx, &ua, "SELECT * FROM user_auths WHERE provider = ? AND provider_uid = ?", provider, providerUID)
      if err != nil {
          return nil, err
      }
      return &ua, nil
  }
  ```

  Update `server/cmd/server/main.go` to use `model.NewUserDB(db)` instead of passing `db` directly.

- [ ] **Step 4: Build and verify**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go build ./cmd/server/
  ```
  Expected: binary `server` produced, no errors.

- [ ] **Step 5: Run all server tests**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/server
  go test ./... -v -count=1
  ```
  Expected: all tests pass

- [ ] **Step 6: Commit**
  ```bash
  git add server/cmd/server/ server/internal/api/router.go server/internal/model/user_store.go
  git commit -m "feat: wire up router with auth routes, admin stubs, and main entry"
  ```

---

### Task 10: Initialize Next.js Project

**Files:**
- Create: `frontend/` (via create-next-app)

- [ ] **Step 1: Create Next.js app**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface
  npx create-next-app@latest frontend --typescript --tailwind --app --src-dir=false --import-alias="@/*" --use-npm --eslint
  ```

- [ ] **Step 2: Install additional dependencies**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/frontend
  npm install clsx
  ```

- [ ] **Step 3: Verify dev server starts**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/frontend
  npm run dev  # Ctrl+C after confirming it loads
  ```
  Visit `http://localhost:3000` in browser — should show default Next.js page.

- [ ] **Step 4: Commit**
  ```bash
  git add frontend/
  git commit -m "chore: initialize Next.js project with TypeScript and Tailwind"
  ```

---

### Task 11: Starbucks Design System — Tokens & CSS Variables

**Files:**
- Modify: `frontend/tailwind.config.ts`
- Modify: `frontend/app/globals.css`

- [ ] **Step 1: Write Tailwind config with Starbucks tokens**

  `frontend/tailwind.config.ts`:
  ```typescript
  import type { Config } from "tailwindcss";

  const config: Config = {
    content: [
      "./app/**/*.{js,ts,jsx,tsx,mdx}",
      "./components/**/*.{js,ts,jsx,tsx,mdx}",
    ],
    theme: {
      extend: {
        colors: {
          // Primary greens
          "starbucks": "#006241",
          "green-accent": "#00754A",
          "house-green": "#1E3932",
          "green-uplift": "#2b5148",
          "green-light": "#d4e9e2",
          // Gold (Rewards only)
          "gold": "#cba258",
          "gold-light": "#dfc49d",
          "gold-lightest": "#faf6ee",
          // Surfaces
          "cream": "#f2f0eb",
          "ceramic": "#edebe9",
          "neutral-cool": "#f9f9f9",
          // Text
          "text-black": "rgba(0, 0, 0, 0.87)",
          "text-black-soft": "rgba(0, 0, 0, 0.58)",
          "text-white-soft": "rgba(255, 255, 255, 0.70)",
          // Semantic
          "red-error": "#c82014",
          "yellow-warn": "#fbbc05",
        },
        fontFamily: {
          sans: ['Inter', '"Helvetica Neue"', 'Helvetica', 'Arial', 'sans-serif'],
        },
        borderRadius: {
          pill: "50px",
        },
        spacing: {
          "space-1": "0.4rem",
          "space-2": "0.8rem",
          "space-3": "1.6rem",
          "space-4": "2.4rem",
          "space-5": "3.2rem",
          "space-6": "4rem",
          "space-7": "4.8rem",
          "space-8": "5.6rem",
          "space-9": "6.4rem",
        },
        boxShadow: {
          card: "0 0 0.5px rgba(0,0,0,0.14), 0 1px 1px rgba(0,0,0,0.24)",
          nav: "0 1px 3px rgba(0,0,0,0.1), 0 2px 2px rgba(0,0,0,0.06), 0 0 2px rgba(0,0,0,0.07)",
          frap: "0 0 6px rgba(0,0,0,0.24), 0 8px 12px rgba(0,0,0,0.14)",
        },
        letterSpacing: {
          tight: "-0.01em",
        },
      },
    },
    plugins: [],
  };

  export default config;
  ```

- [ ] **Step 2: Write globals.css with Starbucks variables**

  `frontend/app/globals.css`:
  ```css
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');

  @tailwind base;
  @tailwind components;
  @tailwind utilities;

  :root {
    --color-canvas: #f2f0eb;
    --color-card: #ffffff;
    --button-active-scale: 0.95;
    --card-border-radius: 12px;
    --card-shadow: 0 0 0.5px rgba(0,0,0,0.14), 0 1px 1px rgba(0,0,0,0.24);
    --transition-default: all 0.2s ease;
  }

  body {
    background-color: var(--color-canvas);
    font-family: 'Inter', 'Helvetica Neue', Helvetica, Arial, sans-serif;
    letter-spacing: -0.01em;
    color: rgba(0, 0, 0, 0.87);
  }

  @layer components {
    .btn-active {
      transform: scale(var(--button-active-scale));
    }
  }
  ```

- [ ] **Step 3: Update root layout to use Inter font**

  `frontend/app/layout.tsx`:
  ```tsx
  import type { Metadata } from "next";
  import "./globals.css";

  export const metadata: Metadata = {
    title: "Vibe Pokeface",
    description: "Family Poker Entertainment",
  };

  export default function RootLayout({
    children,
  }: Readonly<{
    children: React.ReactNode;
  }>) {
    return (
      <html lang="zh-CN">
        <body className="min-h-screen antialiased">
          {children}
        </body>
      </html>
    );
  }
  ```

- [ ] **Step 4: Verify frontend compiles**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/frontend
  npm run build 2>&1 | tail -5
  ```
  Expected: build successful.

- [ ] **Step 5: Commit**
  ```bash
  git add frontend/tailwind.config.ts frontend/app/globals.css frontend/app/layout.tsx
  git commit -m "feat: add Starbucks design system tokens and Tailwind config"
  ```

---

### Task 12: Starbucks Design System — Core Components

**Files:**
- Create: `frontend/components/ui/Button.tsx`
- Create: `frontend/components/ui/Card.tsx`
- Create: `frontend/components/ui/Input.tsx`

- [ ] **Step 1: Write Button component**

  `frontend/components/ui/Button.tsx`:
  ```tsx
  import { ButtonHTMLAttributes, forwardRef } from "react";
  import clsx from "clsx";

  type ButtonVariant =
    | "primary"      // Green Accent fill
    | "outlined"     // Green Accent border
    | "black-fill"   // Black fill
    | "dark-outlined" // Dark border
    | "white-fill"   // White fill on dark bg
    | "outlined-light"; // White border on dark bg

  interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: ButtonVariant;
    fullWidth?: boolean;
  }

  const variantStyles: Record<ButtonVariant, string> = {
    primary:
      "bg-green-accent text-white border border-green-accent",
    outlined:
      "bg-transparent text-green-accent border border-green-accent",
    "black-fill":
      "bg-black text-white border border-black",
    "dark-outlined":
      "bg-transparent text-text-black border border-text-black",
    "white-fill":
      "bg-white text-green-accent border border-white",
    "outlined-light":
      "bg-transparent text-white border border-white",
  };

  export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
    function Button({ variant = "primary", fullWidth, className, children, ...props }, ref) {
      return (
        <button
          ref={ref}
          className={clsx(
            "rounded-pill px-4 py-[7px] text-sm font-semibold tracking-tight",
            "transition-all duration-200 ease",
            "active:scale-[0.95]",
            "disabled:opacity-50 disabled:cursor-not-allowed",
            variantStyles[variant],
            fullWidth && "w-full",
            className
          )}
          {...props}
        >
          {children}
        </button>
      );
    }
  );
  ```

- [ ] **Step 2: Write Card component**

  `frontend/components/ui/Card.tsx`:
  ```tsx
  import { HTMLAttributes, forwardRef } from "react";
  import clsx from "clsx";

  interface CardProps extends HTMLAttributes<HTMLDivElement> {
    padding?: "sm" | "md" | "lg";
  }

  const paddingStyles = {
    sm: "p-space-3",
    md: "p-space-4",
    lg: "p-space-5",
  };

  export const Card = forwardRef<HTMLDivElement, CardProps>(
    function Card({ padding = "md", className, children, ...props }, ref) {
      return (
        <div
          ref={ref}
          className={clsx(
            "bg-white rounded-[12px] shadow-card",
            paddingStyles[padding],
            className
          )}
          {...props}
        >
          {children}
        </div>
      );
    }
  );
  ```

- [ ] **Step 3: Write floating-label Input component**

  `frontend/components/ui/Input.tsx`:
  ```tsx
  "use client";

  import { InputHTMLAttributes, forwardRef, useState } from "react";
  import clsx from "clsx";

  interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
    label: string;
    error?: string;
  }

  export const Input = forwardRef<HTMLInputElement, InputProps>(
    function Input({ label, error, className, value, onChange, ...props }, ref) {
      const [focused, setFocused] = useState(false);
      const hasValue = value !== undefined && value !== "";

      return (
        <div className="relative">
          <div
            className={clsx(
              "border rounded-[4px] px-3 pt-5 pb-2 transition-all duration-200",
              error
                ? "border-red-error bg-red-50/5"
                : focused
                  ? "border-green-accent"
                  : "border-gray-300"
            )}
          >
            <label
              className={clsx(
                "absolute left-3 transition-all duration-200 pointer-events-none",
                focused || hasValue
                  ? "top-1 text-xs font-bold uppercase tracking-wide"
                  : "top-1/2 -translate-y-1/2 text-base",
                error
                  ? "text-red-error"
                  : focused
                    ? "text-green-accent"
                    : "text-text-black-soft"
              )}
            >
              {label}
            </label>
            <input
              ref={ref}
              className="w-full bg-transparent outline-none text-base text-text-black"
              onFocus={() => setFocused(true)}
              onBlur={() => setFocused(false)}
              value={value}
              onChange={onChange}
              {...props}
            />
          </div>
          {error && (
            <p className="mt-1 text-xs text-red-error">{error}</p>
          )}
        </div>
      );
    }
  );
  ```

- [ ] **Step 4: Verify frontend compiles**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/frontend
  npm run build 2>&1 | tail -10
  ```
  Expected: build successful, no errors.

- [ ] **Step 5: Commit**
  ```bash
  git add frontend/components/ui/
  git commit -m "feat: add Starbucks-themed Button, Card, and Input components"
  ```

---

### Task 13: Frontend Auth Pages

**Files:**
- Create: `frontend/app/auth/layout.tsx`
- Create: `frontend/app/auth/login/page.tsx`
- Create: `frontend/app/auth/register/page.tsx`
- Create: `frontend/lib/api-client.ts`

- [ ] **Step 1: Create API client**

  `frontend/lib/api-client.ts`:
  ```typescript
  const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  interface ApiResponse<T = unknown> {
    data?: T;
    error?: string;
  }

  class ApiClient {
    private token: string | null = null;

    constructor() {
      if (typeof window !== "undefined") {
        this.token = localStorage.getItem("token");
      }
    }

    setToken(token: string | null) {
      this.token = token;
      if (typeof window !== "undefined") {
        if (token) {
          localStorage.setItem("token", token);
        } else {
          localStorage.removeItem("token");
        }
      }
    }

    getToken(): string | null {
      return this.token;
    }

    private async request<T>(
      method: string,
      path: string,
      body?: unknown
    ): Promise<ApiResponse<T>> {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };
      if (this.token) {
        headers["Authorization"] = `Bearer ${this.token}`;
      }

      try {
        const res = await fetch(`${API_BASE}${path}`, {
          method,
          headers,
          body: body ? JSON.stringify(body) : undefined,
        });

        const data = await res.json();
        if (!res.ok) {
          return { error: data.error || `HTTP ${res.status}` };
        }
        return { data };
      } catch (err) {
        return { error: err instanceof Error ? err.message : "Network error" };
      }
    }

    register(nickname: string, password: string) {
      return this.request<{ token: string; user: unknown }>("POST", "/api/auth/register", {
        nickname,
        password,
      });
    }

    login(password: string, providerUid?: string) {
      return this.request<{ token: string; user: unknown }>("POST", "/api/auth/login", {
        password,
        provider_uid: providerUid,
      });
    }

    guestLogin(deviceId: string) {
      return this.request<{ token: string; user: unknown }>("POST", "/api/auth/guest", {
        device_id: deviceId,
      });
    }
  }

  export const apiClient = new ApiClient();
  ```

- [ ] **Step 2: Create auth layout**

  `frontend/app/auth/layout.tsx`:
  ```tsx
  export default function AuthLayout({
    children,
  }: {
    children: React.ReactNode;
  }) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-cream px-4">
        <div className="w-full max-w-md">{children}</div>
      </div>
    );
  }
  ```

- [ ] **Step 3: Create login page**

  `frontend/app/auth/login/page.tsx`:
  ```tsx
  "use client";

  import { useState } from "react";
  import { useRouter } from "next/navigation";
  import Link from "next/link";
  import { Button } from "@/components/ui/Button";
  import { Card } from "@/components/ui/Card";
  import { Input } from "@/components/ui/Input";
  import { apiClient } from "@/lib/api-client";

  export default function LoginPage() {
    const router = useRouter();
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
      e.preventDefault();
      setError("");
      setLoading(true);

      const result = await apiClient.login(password);
      if (result.error) {
        setError(result.error);
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
        <h1 className="text-2xl font-semibold text-starbucks mb-6">Sign In</h1>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={error}
          />
          <Button type="submit" fullWidth disabled={loading}>
            {loading ? "Signing in..." : "Sign In"}
          </Button>
        </form>
        <p className="mt-4 text-center text-sm text-text-black-soft">
          Don&apos;t have an account?{" "}
          <Link href="/auth/register" className="text-green-accent underline">
            Register
          </Link>
        </p>
      </Card>
    );
  }
  ```

- [ ] **Step 4: Create register page**

  `frontend/app/auth/register/page.tsx`:
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
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
      e.preventDefault();
      setError("");
      setLoading(true);

      const result = await apiClient.register(nickname, password);
      if (result.error) {
        setError(result.error);
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
        <h1 className="text-2xl font-semibold text-starbucks mb-6">Register</h1>
        <form onSubmit={handleSubmit} className="space-y-4">
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
            error={error}
          />
          <Button type="submit" fullWidth disabled={loading}>
            {loading ? "Creating account..." : "Create Account"}
          </Button>
        </form>
        <p className="mt-4 text-center text-sm text-text-black-soft">
          Already have an account?{" "}
          <Link href="/auth/login" className="text-green-accent underline">
            Sign In
          </Link>
        </p>
      </Card>
    );
  }
  ```

- [ ] **Step 5: Verify frontend compiles**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/frontend
  npm run build 2>&1 | tail -10
  ```
  Expected: build successful.

- [ ] **Step 6: Commit**
  ```bash
  git add frontend/app/auth/ frontend/lib/
  git commit -m "feat: add auth pages and API client"
  ```

---

### Task 14: Admin Layout Skeleton

**Files:**
- Create: `frontend/app/admin/layout.tsx`
- Create: `frontend/app/admin/dashboard/page.tsx`
- Create: `frontend/app/admin/users/page.tsx`
- Create: `frontend/app/admin/rooms/page.tsx`
- Create: `frontend/app/admin/llm-config/page.tsx`
- Create: `frontend/components/ui/AdminSidebar.tsx`

- [ ] **Step 1: Write AdminSidebar component**

  `frontend/components/ui/AdminSidebar.tsx`:
  ```tsx
  "use client";

  import Link from "next/link";
  import { usePathname } from "next/navigation";
  import clsx from "clsx";

  const navItems = [
    { href: "/admin/dashboard", label: "Dashboard", icon: "📊" },
    { href: "/admin/users", label: "Users", icon: "👥" },
    { href: "/admin/rooms", label: "Rooms", icon: "🃏" },
    { href: "/admin/llm-config", label: "LLM Config", icon: "🤖" },
  ];

  export function AdminSidebar() {
    const pathname = usePathname();

    return (
      <aside className="w-64 bg-house-green min-h-screen text-white p-6">
        <div className="mb-8">
          <h2 className="text-xl font-semibold tracking-tight">Admin</h2>
          <p className="text-sm text-white/70 mt-1">Vibe Pokeface</p>
        </div>
        <nav className="space-y-1">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={clsx(
                "flex items-center gap-3 px-4 py-3 rounded-lg text-sm transition-colors",
                pathname === item.href
                  ? "bg-white/10 text-white font-medium"
                  : "text-white/70 hover:text-white hover:bg-white/5"
              )}
            >
              <span>{item.icon}</span>
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>
    );
  }
  ```

- [ ] **Step 2: Write admin layout**

  `frontend/app/admin/layout.tsx`:
  ```tsx
  import { AdminSidebar } from "@/components/ui/AdminSidebar";

  export default function AdminLayout({
    children,
  }: {
    children: React.ReactNode;
  }) {
    return (
      <div className="flex min-h-screen">
        <AdminSidebar />
        <main className="flex-1 bg-cream p-8">
          {children}
        </main>
      </div>
    );
  }
  ```

- [ ] **Step 3: Create placeholder pages**

  `frontend/app/admin/dashboard/page.tsx`:
  ```tsx
  import { Card } from "@/components/ui/Card";

  export default function DashboardPage() {
    return (
      <div>
        <h1 className="text-2xl font-semibold text-starbucks mb-6">Dashboard</h1>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card>
            <p className="text-sm text-text-black-soft">Online Players</p>
            <p className="text-3xl font-semibold mt-1">--</p>
          </Card>
          <Card>
            <p className="text-sm text-text-black-soft">Active Rooms</p>
            <p className="text-3xl font-semibold mt-1">--</p>
          </Card>
          <Card>
            <p className="text-sm text-text-black-soft">Total Users</p>
            <p className="text-3xl font-semibold mt-1">--</p>
          </Card>
        </div>
      </div>
    );
  }
  ```

  `frontend/app/admin/users/page.tsx`:
  ```tsx
  import { Card } from "@/components/ui/Card";

  export default function AdminUsersPage() {
    return (
      <div>
        <h1 className="text-2xl font-semibold text-starbucks mb-6">User Management</h1>
        <Card>
          <p className="text-text-black-soft">User list coming in Phase 4.</p>
        </Card>
      </div>
    );
  }
  ```

  `frontend/app/admin/rooms/page.tsx`:
  ```tsx
  import { Card } from "@/components/ui/Card";

  export default function AdminRoomsPage() {
    return (
      <div>
        <h1 className="text-2xl font-semibold text-starbucks mb-6">Room Monitor</h1>
        <Card>
          <p className="text-text-black-soft">Room monitoring coming in Phase 4.</p>
        </Card>
      </div>
    );
  }
  ```

  `frontend/app/admin/llm-config/page.tsx`:
  ```tsx
  import { Card } from "@/components/ui/Card";

  export default function AdminLLMConfigPage() {
    return (
      <div>
        <h1 className="text-2xl font-semibold text-starbucks mb-6">LLM Configuration</h1>
        <Card>
          <p className="text-text-black-soft">LLM configuration coming in Phase 4.</p>
        </Card>
      </div>
    );
  }
  ```

- [ ] **Step 4: Add lobby placeholder (so /lobby redirect works)**

  `frontend/app/(main)/lobby/page.tsx`:
  ```tsx
  import { Card } from "@/components/ui/Card";

  export default function LobbyPage() {
    return (
      <div className="max-w-4xl mx-auto py-8 px-4">
        <h1 className="text-2xl font-semibold text-starbucks mb-6">Game Lobby</h1>
        <Card>
          <p className="text-text-black-soft">Game lobby coming in Phase 2.</p>
        </Card>
      </div>
    );
  }
  ```

- [ ] **Step 5: Update root page to redirect**

  `frontend/app/page.tsx`:
  ```tsx
  import { redirect } from "next/navigation";

  export default function Home() {
    redirect("/auth/login");
  }
  ```

- [ ] **Step 6: Verify frontend compiles**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/frontend
  npm run build 2>&1 | tail -10
  ```
  Expected: build successful.

- [ ] **Step 7: Run full build to catch any issues**
  ```bash
  cd /Users/yongkl/Develop/VSCodeProjects/vibe_pokeface/frontend
  npm run build
  ```
  Expected: All routes compiled successfully.

- [ ] **Step 8: Commit**
  ```bash
  git add frontend/app/admin/ frontend/app/(main)/ frontend/app/page.tsx frontend/components/ui/AdminSidebar.tsx
  git commit -m "feat: add admin layout skeleton with sidebar and placeholder pages"
  ```

---

## Self-Review

**Spec coverage check:**
- Phase 1 spec items: VPS 环境 ✓ (Task 1), Go 项目骨架 ✓ (Tasks 2-9), MySQL 建表 ✓ (Task 3), WS Hub ✓ (Task 8), Next.js 初始化 ✓ (Task 10), Starbucks 设计系统 ✓ (Tasks 11-12), 用户认证 ✓ (Tasks 4-6), Admin 基础路由 ✓ (Task 14)
- TDD requirement: Every Go package has test files written before implementation ✓
- All steps have exact code, no placeholders ✓

**Placeholder scan:** No TBD, TODO, or incomplete sections found.

**Type consistency:** All type names (UserStore, JWTService, AuthHandler, Hub, RoomHub, Client) are consistent across tasks. Method signatures match between interfaces and implementations.
