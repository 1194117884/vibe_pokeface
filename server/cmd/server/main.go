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
	"github.com/yongkl/vibe-pokeface/internal/api/admin"
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

	userDB := model.NewUserDB(db)
	jwtSvc := auth.NewJWTService(cfg.JWTSecret)

	gameStore := model.NewGameStore(db)
	aiStore := model.NewAIStore(db)

	hub := ws.NewHub(gameStore, aiStore)
	go hub.Run()

	adminHandler := admin.NewHandler(userDB, gameStore, aiStore, hub)

	lkConfig := api.LiveKitConfig{
		APIKey:    os.Getenv("LIVEKIT_API_KEY"),
		APISecret: os.Getenv("LIVEKIT_API_SECRET"),
		Host:      os.Getenv("LIVEKIT_HOST"),
	}

	router := api.NewRouter(userDB, jwtSvc, hub, middleware.CORSConfig{
		AllowedOrigins: cfg.AllowedOrigins,
	}, lkConfig, adminHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

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
