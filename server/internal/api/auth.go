package api

import (
	"encoding/json"
	"net/http"

	"github.com/yongkl/vibe-pokeface/internal/auth"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type AuthHandler struct {
	store model.UserStore
	jwt   *auth.JWTService
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

	if req.ProviderUID == "" {
		http.Error(w, `{"error":"nickname required"}`, http.StatusBadRequest)
		return
	}
	providerUID := req.ProviderUID

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
		token, _ := h.jwt.GenerateToken(existing.ID, existing.Role)
		json.NewEncoder(w).Encode(authResponse{Token: token, User: existing})
		return
	}

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
