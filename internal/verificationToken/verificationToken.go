package verificationToken

import (
	"crypto/rand"
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
