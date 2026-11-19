package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	_, error := rand.Read(bytes)
	if error != nil {
		return "", error
	}
	return hex.EncodeToString(bytes), nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}