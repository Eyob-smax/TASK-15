package security_test

import (
	"testing"

	"fitcommerce/internal/security"
)

func TestHashPassword_RoundTrip(t *testing.T) {
	hash, salt, err := security.HashPassword("correcthorsebatterystaple")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hash == "" || salt == "" {
		t.Fatal("HashPassword returned empty hash or salt")
	}

	ok, err := security.VerifyPassword("correcthorsebatterystaple", hash, salt)
	if err != nil {
		t.Fatalf("VerifyPassword error: %v", err)
	}
	if !ok {
		t.Fatal("expected VerifyPassword to return true for correct password")
	}
}

func TestHashPassword_WrongPasswordRejected(t *testing.T) {
	hash, salt, err := security.HashPassword("correctpassword")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}

	ok, err := security.VerifyPassword("wrongpassword", hash, salt)
	if err != nil {
		t.Fatalf("VerifyPassword error: %v", err)
	}
	if ok {
		t.Fatal("expected VerifyPassword to return false for wrong password")
	}
}

func TestHashPassword_DifferentSaltsProduceDifferentHashes(t *testing.T) {
	const password = "samepassword"
	hash1, salt1, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("first HashPassword error: %v", err)
	}
	hash2, salt2, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("second HashPassword error: %v", err)
	}

	if salt1 == salt2 {
		t.Error("expected different salts for two HashPassword calls")
	}
	if hash1 == hash2 {
		t.Error("expected different hashes when salts differ")
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, salt, err := security.HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword(empty) error: %v", err)
	}

	ok, err := security.VerifyPassword("", hash, salt)
	if err != nil {
		t.Fatalf("VerifyPassword(empty) error: %v", err)
	}
	if !ok {
		t.Fatal("expected VerifyPassword to accept empty password when hashed as empty")
	}

	ok2, _ := security.VerifyPassword("notempty", hash, salt)
	if ok2 {
		t.Fatal("expected VerifyPassword to reject non-empty password against empty-password hash")
	}
}
