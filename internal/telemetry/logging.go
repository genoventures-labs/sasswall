package telemetry

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"
)

type Event struct {
	Timestamp        time.Time `json:"ts"`
	IP               string    `json:"ip"`
	UA               string    `json:"ua"`
	Host             string    `json:"host"`
	Path             string    `json:"path"`
	Category         string    `json:"category"`
	Persona          string    `json:"persona"`
	Strikes          int       `json:"strikes"`
	Denied           bool      `json:"denied"`
	DenyUntil        string    `json:"deny_until,omitempty"`
	Limited          bool      `json:"limited"`
	SessionID        string    `json:"session_id,omitempty"`
	SequenceScore    int       `json:"sequence_score,omitempty"`
	ProfileID        string    `json:"profile_id,omitempty"`
	DeceptionVariant string    `json:"deception_variant,omitempty"`
	ChallengeIssued  bool      `json:"challenge_issued"`
	DecoySuccess     bool      `json:"decoy_success"`
	CanaryTokenID    string    `json:"canary_token_id,omitempty"`
	FairnessStep     int       `json:"fairness_step,omitempty"`
	LatencyMS        int64     `json:"latency_ms"`
}

type Logger struct {
	mu sync.Mutex
	l  *log.Logger
}

func New(w io.Writer) *Logger {
	return &Logger{l: log.New(w, "", 0)}
}

func (j *Logger) Emit(ev Event) {
	j.mu.Lock()
	defer j.mu.Unlock()
	b, _ := json.Marshal(ev)
	j.l.Println(string(b))
}
