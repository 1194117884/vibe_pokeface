package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yongkl/vibe-pokeface/internal/api/admin"
	"github.com/yongkl/vibe-pokeface/internal/api/middleware"
	"github.com/yongkl/vibe-pokeface/internal/api/ws"
	"github.com/yongkl/vibe-pokeface/internal/auth"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

func NewRouter(store model.UserStore, jwt *auth.JWTService, hub *ws.Hub, corsCfg middleware.CORSConfig, lkConfig LiveKitConfig, adminHandler *admin.Handler, roomHandler *RoomHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logging)
	r.Use(middleware.CORS(corsCfg))

	authHandler := NewAuthHandler(store, jwt)
	authRateLimiter := middleware.NewRateLimiter(10, time.Second)

	r.Route("/api/auth", func(r chi.Router) {
		r.Use(authRateLimiter.Middleware)
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/guest", authHandler.GuestLogin)
	})

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/ws", hub.HandleWS)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwt))

		r.Group(func(r chi.Router) {
			r.Use(middleware.AdminOnly)
			r.Route("/api/admin", func(r chi.Router) {
				r.Get("/dashboard", adminHandler.Dashboard.ServeHTTP)
				r.Get("/users", adminHandler.Users.List)
				r.Put("/users/{id}/status", adminHandler.Users.UpdateStatus)
				r.Get("/rooms", adminHandler.Rooms.List)
				r.Get("/rooms/{id}", adminHandler.Rooms.Get)
				r.Get("/ai-characters", adminHandler.AIChars.List)
				r.Post("/ai-characters", adminHandler.AIChars.Create)
				r.Put("/ai-characters/{id}", adminHandler.AIChars.Update)
				r.Delete("/ai-characters/{id}", adminHandler.AIChars.Delete)
				r.Get("/llm-configs", adminHandler.LLMConfig.List)
				r.Post("/llm-configs", adminHandler.LLMConfig.Create)
				r.Delete("/llm-configs/{id}", adminHandler.LLMConfig.Delete)
				r.Get("/llm-stats", adminHandler.LLMConfig.GetStats)
				r.Get("/scores", adminHandler.Scores.GetBalance)
				r.Post("/scores/adjust", adminHandler.Scores.Adjust)
			})
		})

		r.Get("/api/livekit/token", LiveKitTokenHandler(lkConfig))

		r.Get("/api/room/{id}/reconnect", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message":"reconnect - coming soon","status":"ok"}`))
		})

		r.Get("/api/ai-characters", adminHandler.AIChars.ListEnabled)
		r.Get("/api/rooms", roomHandler.ListRooms)
		r.Post("/api/rooms", roomHandler.CreateRoom)
	})

	return r
}
