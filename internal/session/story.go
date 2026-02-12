package session

import (
	"crypto/sha1"
	"fmt"
	"sync"
	"time"
)

type Phase string

const (
	PhaseObserve Phase = "observe"
	PhaseEngage  Phase = "engage"
	PhaseSink    Phase = "sink"
)

type Story struct {
	SessionID      string
	Phase          Phase
	Score          int
	LastSeen       time.Time
	DelayedSuccess int
	Requests       int
}

type Store struct {
	mu    sync.Mutex
	items map[string]*Story
	ttl   time.Duration
}

func NewStore(ttl time.Duration) *Store {
	return &Store{items: map[string]*Story{}, ttl: ttl}
}

func SessionKey(ip, ua string, now time.Time) string {
	bucket := now.Unix() / int64((15 * time.Minute).Seconds())
	h := sha1.Sum([]byte(ua))
	return fmt.Sprintf("%s:%x:%d", ip, h[:6], bucket)
}

func (s *Store) Update(key string, scoreDelta int, now time.Time) *Story {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gc(now)
	st, ok := s.items[key]
	if !ok {
		st = &Story{SessionID: key, Phase: PhaseObserve}
		s.items[key] = st
	}
	st.Score += scoreDelta
	st.Requests++
	st.LastSeen = now
	switch {
	case st.Score >= 8:
		st.Phase = PhaseSink
	case st.Score >= 4:
		st.Phase = PhaseEngage
	default:
		st.Phase = PhaseObserve
	}
	copy := *st
	return &copy
}

func (s *Store) MarkDelayedSuccess(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.items[key]; ok {
		st.DelayedSuccess++
	}
}

func (s *Store) Current(key string) Story {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.items[key]; ok {
		cp := *st
		return cp
	}
	return Story{SessionID: key, Phase: PhaseObserve}
}

func (s *Store) gc(now time.Time) {
	for k, st := range s.items {
		if now.Sub(st.LastSeen) > s.ttl {
			delete(s.items, k)
		}
	}
}
