package escalation

import (
	"sync"
	"time"
)

type entry struct {
	timestamps []time.Time
	denyUntil  time.Time
}

type Tracker struct {
	mu        sync.Mutex
	entries   map[string]*entry
	window    time.Duration
	threshold int
	denyFor   time.Duration
}

func New(window time.Duration, threshold int, denyFor time.Duration) *Tracker {
	return &Tracker{entries: make(map[string]*entry), window: window, threshold: threshold, denyFor: denyFor}
}

func (t *Tracker) State(ip string, now time.Time) (strikes int, denied bool, denyUntil time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e := t.get(ip)
	e.prune(now.Add(-t.window))
	return len(e.timestamps), now.Before(e.denyUntil), e.denyUntil
}

func (t *Tracker) AddStrike(ip string, now time.Time) (strikes int, denied bool, denyUntil time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e := t.get(ip)
	e.prune(now.Add(-t.window))
	e.timestamps = append(e.timestamps, now)
	if len(e.timestamps) >= t.threshold {
		e.denyUntil = now.Add(t.denyFor)
	}
	return len(e.timestamps), now.Before(e.denyUntil), e.denyUntil
}

func (t *Tracker) get(ip string) *entry {
	e, ok := t.entries[ip]
	if !ok {
		e = &entry{}
		t.entries[ip] = e
	}
	return e
}

func (e *entry) prune(cutoff time.Time) {
	if len(e.timestamps) == 0 {
		return
	}
	out := e.timestamps[:0]
	for _, ts := range e.timestamps {
		if ts.After(cutoff) {
			out = append(out, ts)
		}
	}
	e.timestamps = out
}
