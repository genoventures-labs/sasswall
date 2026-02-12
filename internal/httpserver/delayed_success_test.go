package httpserver

import (
	"testing"

	"sasswall/internal/classify"
	"sasswall/internal/config"
	"sasswall/internal/session"
)

func TestDelayedSuccessNeverNormal(t *testing.T) {
	cfg := config.Default()
	cfg.Deception.DelayedSuccess.Enabled = true
	story := &session.Story{Requests: 10, DelayedSuccess: 0}
	if shouldDelayedSuccess(&cfg, classify.CategoryNormal, "/.env", story) {
		t.Fatalf("normal category should never receive delayed success")
	}
}
