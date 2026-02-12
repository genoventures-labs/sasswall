package session

import "testing"

func TestPresenceForScoreBoundaries(t *testing.T) {
	if got := PresenceForScore(1, 4, 8, true); got != PresenceObserve {
		t.Fatalf("expected observe, got %s", got)
	}
	if got := PresenceForScore(4, 4, 8, true); got != PresenceLockOn {
		t.Fatalf("expected lock-on, got %s", got)
	}
	if got := PresenceForScore(8, 4, 8, true); got != PresencePressure {
		t.Fatalf("expected pressure, got %s", got)
	}
	if got := PresenceForScore(100, 4, 8, false); got != "" {
		t.Fatalf("expected empty state for non-hostile, got %s", got)
	}
}

func TestPresenceTransitionNames(t *testing.T) {
	if got := PresenceTransition(PresenceObserve, PresenceLockOn); got != "observe_to_lock_on" {
		t.Fatalf("unexpected transition: %s", got)
	}
	if got := PresenceTransition(PresenceLockOn, PresencePressure); got != "lock_on_to_pressure" {
		t.Fatalf("unexpected transition: %s", got)
	}
}
