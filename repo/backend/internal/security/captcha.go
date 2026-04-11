package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"math/big"
	"strings"
)

const captchaSaltSize = 16

// GenerateMathCaptcha creates a simple addition problem as an on-device CAPTCHA
// challenge requiring no external service. Returns the human-readable challenge
// string and the expected string answer.
func GenerateMathCaptcha() (challengeData string, answer string, err error) {
	x, err := cryptoRandInt(20)
	if err != nil {
		return "", "", fmt.Errorf("generate captcha x: %w", err)
	}
	y, err := cryptoRandInt(20)
	if err != nil {
		return "", "", fmt.Errorf("generate captcha y: %w", err)
	}

	challengeData = fmt.Sprintf("What is %d + %d?", x+1, y+1)
	answer = fmt.Sprintf("%d", x+y+2)
	return challengeData, answer, nil
}

// GenerateCaptchaSalt creates a random salt used to derive a one-way stored
// verification hash for a CAPTCHA answer.
func GenerateCaptchaSalt() ([]byte, error) {
	salt := make([]byte, captchaSaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate captcha salt: %w", err)
	}
	return salt, nil
}

// HashCaptchaAnswer derives the stored verification hash for a CAPTCHA answer.
func HashCaptchaAnswer(answer string, salt []byte) []byte {
	normalized := NormalizeCaptchaAnswer(answer)
	input := append([]byte(normalized+":"), salt...)
	sum := sha256.Sum256(input)
	return sum[:]
}

// VerifyCaptchaAnswer checks whether the submitted answer matches the stored
// derived verification hash using constant-time comparison.
func VerifyCaptchaAnswer(storedHash, salt []byte, submitted string) bool {
	if len(storedHash) == 0 || len(salt) == 0 {
		return false
	}
	submittedHash := HashCaptchaAnswer(submitted, salt)
	return subtle.ConstantTimeCompare(storedHash, submittedHash) == 1
}

// NormalizeCaptchaAnswer canonicalizes a CAPTCHA answer before hashing.
func NormalizeCaptchaAnswer(answer string) string {
	return strings.TrimSpace(strings.ToLower(answer))
}

func cryptoRandInt(maxExclusive int64) (int64, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(maxExclusive))
	if err != nil {
		return 0, err
	}
	return value.Int64(), nil
}
