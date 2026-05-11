package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseDSN    string
	JWTSecret      string
	AllowedOrigins []string
}

func Load() (*Config, error) {
	godotenv.Load() // ignore error if no .env file

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseDSN:    getEnv("DATABASE_DSN", "root:@tcp(127.0.0.1:3306)/pokeface?parseTime=true"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-in-production"),
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
