package ratelimit

import (
	"fmt"
	"sync"
	"time"

	"sasswall/internal/classify"
)

type bucket struct {
	tokens   float64
	lastSeen time.Time
	step     int
}

type Config struct {
	Burst        int
	NormalRPM    int
	HoneyRPM     int
	ScannerRPM   int
	DeniedRPM    int
	Adaptive     bool
	DegradeMax   int
	RecoverAfter time.Duration
}

type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	cfg     Config
}

type Decision struct {
	Allowed    bool
	RetryAfter int
	Step       int
}

func New(cfg Config) *Limiter {
	return &Limiter{cfg: cfg, buckets: map[string]*bucket{}}
}

func (l *Limiter) Allow(ip string, category classify.Category, score int, now time.Time) Decision {
	l.mu.Lock()
	defer l.mu.Unlock()
	k := fmt.Sprintf("%s:%s", ip, category)
	b, ok := l.buckets[k]
	if !ok {
		b = &bucket{tokens: float64(l.cfg.Burst), lastSeen: now}
		l.buckets[k] = b
	}

	rpm := l.rpmFor(category)
	elapsed := now.Sub(b.lastSeen).Minutes()
	if elapsed > 0 {
		b.tokens += elapsed * float64(rpm)
		if b.tokens > float64(l.cfg.Burst) {
			b.tokens = float64(l.cfg.Burst)
		}
	}
	if l.cfg.Adaptive {
		if score >= 8 && b.step < l.cfg.DegradeMax {
			b.step++
		}
		if now.Sub(b.lastSeen) > l.cfg.RecoverAfter && b.step > 0 {
			b.step--
		}
	}
	b.lastSeen = now

	effectiveCost := 1.0
	switch b.step {
	case 1:
		effectiveCost = 1.3
	case 2:
		effectiveCost = 1.8
	case 3:
		effectiveCost = 2.2
	}
	if b.tokens >= effectiveCost {
		b.tokens -= effectiveCost
		return Decision{Allowed: true, Step: b.step}
	}
	retry := int(60 / max(rpm, 1))
	if b.step >= 2 {
		retry += 2
	}
	return Decision{Allowed: false, RetryAfter: retry, Step: b.step}
}

func (l *Limiter) rpmFor(category classify.Category) int {
	switch category {
	case classify.CategoryHoney:
		return l.cfg.HoneyRPM
	case classify.CategoryScanner:
		return l.cfg.ScannerRPM
	case classify.CategoryDenied:
		return l.cfg.DeniedRPM
	default:
		return l.cfg.NormalRPM
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
