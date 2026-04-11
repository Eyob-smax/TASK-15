package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const aesNonceSize = 12 // GCM standard nonce size in bytes

// GenerateAESKey returns a cryptographically random 32-byte key suitable for
// AES-256 encryption.
func GenerateAESKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generating AES key: %w", err)
	}
	return key, nil
}

// EncryptAESGCM encrypts plaintext using AES-256-GCM with a randomly generated
// nonce. The output is: nonce (12 bytes) ∥ ciphertext ∥ GCM tag.
func EncryptAESGCM(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, aesNonceSize)
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptAESGCM decrypts data produced by EncryptAESGCM. The first 12 bytes of
// ciphertext are treated as the GCM nonce.
func DecryptAESGCM(ciphertext, key []byte) ([]byte, error) {
	if len(ciphertext) < aesNonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonce := ciphertext[:aesNonceSize]
	data := ciphertext[aesNonceSize:]

	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting: %w", err)
	}

	return plaintext, nil
}

// DeriveKeyFromRef derives a 32-byte AES-256 key from the given key reference
// string using HKDF-SHA256. IKM = []byte(keyRef), salt = nil,
// info = "fitcommerce-biometric-v1". The derived key is deterministic for a
// given keyRef, so the caller must ensure keyRef is kept secret and rotated
// through the EncryptionKey rotation mechanism rather than changed ad-hoc.
func DeriveKeyFromRef(keyRef string) ([]byte, error) {
	r := hkdf.New(sha256.New, []byte(keyRef), nil, []byte("fitcommerce-biometric-v1"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, fmt.Errorf("DeriveKeyFromRef: %w", err)
	}
	return key, nil
}
