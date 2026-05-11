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
