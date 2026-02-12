package tarpit

import (
	"testing"
	"time"
)

func TestDelayCap(t *testing.T) {
	c := New(Config{DelayMin: 4 * time.Second, DelayMax: 5 * time.Second, HoneyBoost: 2 * time.Second, StrikeBoost: 1 * time.Second})
	d := c.Delay(100, true, 8)
	if d > MaxTotalDelay {
		t.Fatalf("delay exceeded cap: %v", d)
	}
}
