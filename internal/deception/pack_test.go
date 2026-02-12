package deception

import (
	"testing"
	"time"
)

func TestPickWithSignatureDeterministic(t *testing.T) {
	e := NewDefault()
	now := time.Unix(1700000000, 0)
	a := e.PickWithSignature("session-a", "sig-1", 15*time.Minute, now)
	b := e.PickWithSignature("session-a", "sig-1", 15*time.Minute, now)
	if a.ID != b.ID {
		t.Fatalf("expected deterministic pack selection")
	}
}
