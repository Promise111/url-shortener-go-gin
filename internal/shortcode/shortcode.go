package shortcode

import (
	"crypto/rand"
	"errors"
)

const alphabets = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Generate(n int) (string, error) {
	if n < 1 {
		return "", errors.New("n can not be less than 1")
	}

	var b = make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	var out = make([]byte, n)
	for i := range n {
		out[i] = alphabets[int(b[i])%len(alphabets)]
	}
	return string(out), nil
}
