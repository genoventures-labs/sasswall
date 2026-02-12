package canary

import "testing"

func TestTokenDeterministic(t *testing.T) {
	s := New("secret")
	a := s.Token("sess", "/.env", 1)
	b := s.Token("sess", "/.env", 1)
	if a != b {
		t.Fatalf("expected deterministic token")
	}
	if !s.Verify(a, "sess", "/.env", 1) {
		t.Fatalf("expected token verification to pass")
	}
}
