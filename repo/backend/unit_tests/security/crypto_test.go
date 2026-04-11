package security_test

import (
	"bytes"
	"testing"

	"fitcommerce/internal/security"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key, err := security.GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey error: %v", err)
	}

	plaintext := []byte("sensitive biometric payload")
	ciphertext, err := security.EncryptAESGCM(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptAESGCM error: %v", err)
	}

	decrypted, err := security.DecryptAESGCM(ciphertext, key)
	if err != nil {
		t.Fatalf("DecryptAESGCM error: %v", err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted %q != original %q", decrypted, plaintext)
	}
}

func TestDecryptAESGCM_WrongKeyFails(t *testing.T) {
	key1, _ := security.GenerateAESKey()
	key2, _ := security.GenerateAESKey()

	ciphertext, err := security.EncryptAESGCM([]byte("hello"), key1)
	if err != nil {
		t.Fatalf("EncryptAESGCM error: %v", err)
	}

	_, err = security.DecryptAESGCM(ciphertext, key2)
	if err == nil {
		t.Error("expected decryption with wrong key to fail")
	}
}

func TestEncryptAESGCM_DifferentNoncesEachTime(t *testing.T) {
	key, _ := security.GenerateAESKey()
	plaintext := []byte("same plaintext")

	ct1, err := security.EncryptAESGCM(plaintext, key)
	if err != nil {
		t.Fatalf("first EncryptAESGCM error: %v", err)
	}
	ct2, err := security.EncryptAESGCM(plaintext, key)
	if err != nil {
		t.Fatalf("second EncryptAESGCM error: %v", err)
	}
	if bytes.Equal(ct1, ct2) {
		t.Error("expected different ciphertexts (different nonces) for the same plaintext")
	}
}

func TestGenerateAESKey_Length(t *testing.T) {
	key, err := security.GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey error: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("expected AES key length 32, got %d", len(key))
	}
}

func TestDecryptAESGCM_TooShortFails(t *testing.T) {
	key, _ := security.GenerateAESKey()
	_, err := security.DecryptAESGCM([]byte("short"), key)
	if err == nil {
		t.Error("expected error when ciphertext is shorter than nonce size")
	}
}

func TestDeriveKeyFromRef_SameRefSameKey(t *testing.T) {
	key1, err := security.DeriveKeyFromRef("my-secret-ref")
	if err != nil {
		t.Fatalf("DeriveKeyFromRef error: %v", err)
	}
	key2, err := security.DeriveKeyFromRef("my-secret-ref")
	if err != nil {
		t.Fatalf("DeriveKeyFromRef error: %v", err)
	}
	if !bytes.Equal(key1, key2) {
		t.Error("same keyRef should produce same derived key")
	}
}

func TestDeriveKeyFromRef_DifferentRefDifferentKey(t *testing.T) {
	key1, err := security.DeriveKeyFromRef("ref-one")
	if err != nil {
		t.Fatalf("DeriveKeyFromRef error: %v", err)
	}
	key2, err := security.DeriveKeyFromRef("ref-two")
	if err != nil {
		t.Fatalf("DeriveKeyFromRef error: %v", err)
	}
	if bytes.Equal(key1, key2) {
		t.Error("different keyRefs should produce different derived keys")
	}
}

func TestDeriveKeyFromRef_KeyLength(t *testing.T) {
	key, err := security.DeriveKeyFromRef("any-ref")
	if err != nil {
		t.Fatalf("DeriveKeyFromRef error: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("expected derived key length 32, got %d", len(key))
	}
}
