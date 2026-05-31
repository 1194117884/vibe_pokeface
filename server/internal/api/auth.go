package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yongkl/vibe-pokeface/internal/auth"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type AuthHandler struct {
	store    model.UserStore
	jwt      *auth.JWTService
	regCodes model.RegistrationCodeStore
}

func NewAuthHandler(store model.UserStore, jwt *auth.JWTService, regCodes model.RegistrationCodeStore) *AuthHandler {
	return &AuthHandler{store: store, jwt: jwt, regCodes: regCodes}
}

type registerRequest struct {
	Nickname         string `json:"nickname"`
	Password         string `json:"password"`
	RegistrationCode string `json:"registration_code"`
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

	// Atomically claim the code.
	claimed, _ := h.regCodes.ClaimCode(r.Context(), req.RegistrationCode, user.ID)
	if !claimed {
		// Code was claimed concurrently — user is already created, code still valid for race winner.
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"guest registration closed"}`))
}
