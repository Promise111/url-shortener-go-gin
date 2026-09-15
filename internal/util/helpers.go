package util

import (
	"crypto/rand"
	"fmt"
)

func GenerateShortCode(len int) (string, error) {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
