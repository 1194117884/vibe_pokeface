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
