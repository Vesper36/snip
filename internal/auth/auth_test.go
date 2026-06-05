package auth

import (
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("test123")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == "test123" {
		t.Fatal("HashPassword returned plaintext")
	}
}

func TestCheckPassword(t *testing.T) {
	hash, err := HashPassword("test123")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if !CheckPassword("test123", hash) {
		t.Fatal("CheckPassword failed for correct password")
	}
	if CheckPassword("wrong", hash) {
		t.Fatal("CheckPassword passed for wrong password")
	}
}

func TestGenerateJWT(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!!"
	token, err := GenerateJWT(secret, time.Hour, false)
	if err != nil {
		t.Fatalf("GenerateJWT error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateJWT returned empty token")
	}
}

func TestValidateJWT(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!!"
	token, err := GenerateJWT(secret, time.Hour, true)
	if err != nil {
		t.Fatalf("GenerateJWT error: %v", err)
	}

	claims, err := ValidateJWT(secret, token)
	if err != nil {
		t.Fatalf("ValidateJWT error: %v", err)
	}
	if !claims.IsAdmin {
		t.Fatal("Expected IsAdmin=true")
	}
	if claims.Issuer != "snip" {
		t.Fatalf("Expected issuer=snip, got %s", claims.Issuer)
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	_, err := ValidateJWT("secret", "invalid-token")
	if err != ErrInvalidToken {
		t.Fatalf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!!"
	token, err := GenerateJWT(secret, -time.Hour, false) // expired
	if err != nil {
		t.Fatalf("GenerateJWT error: %v", err)
	}
	_, err = ValidateJWT(secret, token)
	if err != ErrInvalidToken {
		t.Fatalf("Expected ErrInvalidToken for expired token, got %v", err)
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!!"
	token, err := GenerateJWT(secret, time.Hour, false)
	if err != nil {
		t.Fatalf("GenerateJWT error: %v", err)
	}
	_, err = ValidateJWT("wrong-secret", token)
	if err != ErrInvalidToken {
		t.Fatalf("Expected ErrInvalidToken for wrong secret, got %v", err)
	}
}

func TestGenerateAPIToken(t *testing.T) {
	token, err := GenerateAPIToken()
	if err != nil {
		t.Fatalf("GenerateAPIToken error: %v", err)
	}
	if len(token) < 60 {
		t.Fatalf("Token too short: %d chars", len(token))
	}
	if token[:5] != "snip_" {
		t.Fatalf("Expected prefix snip_, got %s", token[:5])
	}
}

func TestGenerateAPIToken_Uniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := GenerateAPIToken()
		if err != nil {
			t.Fatalf("GenerateAPIToken error on iteration %d: %v", i, err)
		}
		if tokens[token] {
			t.Fatalf("Duplicate token generated: %s", token)
		}
		tokens[token] = true
	}
}
