package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory      = 64 * 1024
	argon2Iterations  = 3
	argon2Parallelism = 2
	argon2KeyLen      = 32
	argon2SaltLen     = 16
)

// HashPassword hashes the given password using Argon2id with a randomly generated
// salt. Returns base64-encoded hash and salt strings. Both must be stored and
// supplied to VerifyPassword for future verification.
func HashPassword(password string) (hash string, salt string, err error) {
	saltBytes := make([]byte, argon2SaltLen)
	if _, err = rand.Read(saltBytes); err != nil {
		return "", "", fmt.Errorf("generating salt: %w", err)
	}

	hashBytes := argon2.IDKey(
		[]byte(password),
		saltBytes,
		argon2Iterations,
		argon2Memory,
		argon2Parallelism,
		argon2KeyLen,
	)

	return base64.StdEncoding.EncodeToString(hashBytes),
		base64.StdEncoding.EncodeToString(saltBytes),
		nil
}

// VerifyPassword checks whether the given password matches the stored hash and
// salt produced by HashPassword. Uses constant-time comparison to prevent
// timing-based side-channel attacks.
func VerifyPassword(password, storedHash, storedSalt string) (bool, error) {
	saltBytes, err := base64.StdEncoding.DecodeString(storedSalt)
	if err != nil {
		return false, fmt.Errorf("decoding salt: %w", err)
	}

	storedHashBytes, err := base64.StdEncoding.DecodeString(storedHash)
	if err != nil {
		return false, fmt.Errorf("decoding stored hash: %w", err)
	}

	candidateHash := argon2.IDKey(
		[]byte(password),
		saltBytes,
		argon2Iterations,
		argon2Memory,
		argon2Parallelism,
		argon2KeyLen,
	)

	return subtle.ConstantTimeCompare(candidateHash, storedHashBytes) == 1, nil
}
