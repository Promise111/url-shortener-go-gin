package shortcode

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	code, err := Generate(10)

	if err != nil {
		t.Fatalf("Generate(10) unexpected error: %v", err)
	}

	if len(code) != 10 {
		t.Fatalf("Generate(10) len = %d, want 10", len(code))
	}

	for i := 0; i < len(code); i++ {
		if !strings.ContainsRune(alphabets, rune(code[i])){
			t.Fatalf("invalid char %q at %d", code[i], i)
		}
	}
}

func TestGenerateNonPositiveNumber(t *testing.T) {
	_, err := Generate(-2)

	if err == nil {
		t.Fatalf("Generate(-2) expected error, got nil")
	}
}
