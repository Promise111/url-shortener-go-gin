package verificationtoken

import (
	"testing"
)

func TestGenerateVerificationToken(t *testing.T) {
	token, err := GenerateVerificationToken()

	if err != nil {
		t.Fatalf("GenerateVerificationToken() unexpected error: %v", err)
	}

	if len(token) != 64 {
		t.Fatalf("GenerateVerificationToken() len != %d, expects 64", len(token))
	}

	for i := 0; i < len(token); i++ {
		c := token[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("non-hex char %q at %d", c, i)
		}
	}
}
