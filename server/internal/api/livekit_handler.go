package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yongkl/vibe-pokeface/internal/api/middleware"
	"github.com/yongkl/vibe-pokeface/internal/auth"
)

// LiveKitConfig holds the LiveKit server configuration.
type LiveKitConfig struct {
	APIKey    string
	APISecret string
	Host      string
}

// LiveKitTokenResponse is the response for a LiveKit token request.
type LiveKitTokenResponse struct {
	Token   string `json:"token"`
	URL     string `json:"url"`
	Room    string `json:"room"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// LiveKitClaims are the JWT claims for a LiveKit access token.
type LiveKitClaims struct {
	jwt.RegisteredClaims
	Video LiveKitVideoClaims `json:"video"`
}

// LiveKitVideoClaims contains LiveKit-specific video room permissions.
type LiveKitVideoClaims struct {
	Room     string `json:"room"`
	RoomJoin bool   `json:"roomJoin"`
}

// LiveKitTokenHandler generates LiveKit access tokens.
func LiveKitTokenHandler(lkConfig LiveKitConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)
		if !ok || claims == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(LiveKitTokenResponse{
				Success: false,
				Error:   "invalid authentication claims",
			})
			return
		}

		roomID := r.URL.Query().Get("room")
		if roomID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(LiveKitTokenResponse{
				Success: false,
				Error:   "room parameter required",
			})
			return
		}

		identity := formatUserIdentity(claims.UserID)

		// Generate LiveKit token using golang-jwt
		now := time.Now()
		lkClaims := LiveKitClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    lkConfig.APIKey,
				Subject:   identity,
				ExpiresAt: jwt.NewNumericDate(now.Add(4 * time.Hour)),
				NotBefore: jwt.NewNumericDate(now),
				IssuedAt:  jwt.NewNumericDate(now),
			},
			Video: LiveKitVideoClaims{
				Room:     roomID,
				RoomJoin: true,
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, lkClaims)
		signedToken, err := token.SignedString([]byte(lkConfig.APISecret))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(LiveKitTokenResponse{
				Success: false,
				Error:   "failed to generate token: " + err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LiveKitTokenResponse{
			Token:   signedToken,
			URL:     lkConfig.Host,
			Room:    roomID,
			Success: true,
		})
	}
}

// formatUserIdentity formats a user ID for LiveKit participant identity.
func formatUserIdentity(userID int64) string {
	return "user_" + strconv.FormatInt(userID, 10)
}

