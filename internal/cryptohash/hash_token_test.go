package cryptohash

import (
	"testing"
)

func TestHashToken(t *testing.T) {
	token := "1f2ae3c4e5"
	hash := HashToken(token)
	hashRes := "ac725cf1a4869296ede5bacd3dfec213965cef60fccf99ef3cb808dc9d9d0024"

	if token == hash {
		t.Fatalf("HashToken(%v)  hash must not equal token", token)
	}

	if len(hash) != 64 {
		t.Fatalf("HashToken(%v) expects hash length = %d", token, 64)
	}

	if hashRes != hash {
		t.Fatalf("HashToken(%v) not stable", token)
	}

	for i := 0; i < len(hash); i++ {
		c := hash[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("non-hex char %q at %v", c, i)
		}
	}
}

func TestHashTokenEmptyString(t *testing.T) {
	hash := HashToken("")
	empty := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	if hash != empty {
		t.Fatalf("HashToken(\"\") = %q, want %q", hash, empty)
	}
}

func TestHashTokenDifferentInput(t *testing.T) {
	hash1 := HashToken("token-a")
	hash2 := HashToken("token-b")

	if hash1 == hash2 {
		t.Fatalf("HashToken(%q) and HashToken(%q) should differ, both = %q", "token-a", "token-b", hash1)
	}
}
