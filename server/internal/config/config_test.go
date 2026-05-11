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
