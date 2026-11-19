package handler

import (
	"testing"
)

func TestLowerTrim(t *testing.T) {
	cases := []struct{ in, want string }{
		{"  Foo@Bar.COM  ", "foo@bar.com"},
		{"already-clean", "already-clean"},
		{"\tMiXeD\n", "mixed"},
	}

	for _, tc := range cases {
		got := LowerTrim(tc.in)
		if got != tc.want {
			t.Fatalf("LowerTrim(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
