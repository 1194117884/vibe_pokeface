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

	r.Get("/ws", hub.HandleWS)

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

		r.Get("/api/room/{id}/reconnect", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message":"reconnect - coming soon","status":"ok"}`))
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
