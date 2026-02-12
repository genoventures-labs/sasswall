package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"sasswall/internal/config"
	"sasswall/internal/metrics"
	"sasswall/internal/telemetry"
)

func TestHostileRequestGetsSDSAHeaders(t *testing.T) {
	cfg := config.Default()
	cfg.Deception.DelayedSuccess.Enabled = false
	cfg.Tarpit.DelayMin = 1 * time.Millisecond
	cfg.Tarpit.DelayMax = 2 * time.Millisecond
	cfg.Honey.Boost = 0
	srv, err := New(&cfg, telemetry.New(os.Stdout), metrics.New())
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodGet, "http://sasswall.local/.env", nil)
	r.Header.Set("User-Agent", "nmap")
	w := httptest.NewRecorder()
	srv.handle(w, r)
	res := w.Result()

	if res.Header.Get("X-Sasswall-Category") == "" {
		t.Fatalf("missing X-Sasswall-Category")
	}
	if res.Header.Get("X-Sasswall-Profile") == "" {
		t.Fatalf("missing X-Sasswall-Profile")
	}
	if res.Header.Get("X-Sasswall-Narrative") == "" {
		t.Fatalf("missing X-Sasswall-Narrative")
	}
	if res.StatusCode != http.StatusNotFound && res.StatusCode != http.StatusForbidden && res.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("unexpected status code: %d", res.StatusCode)
	}
	if w.Body.Len() == 0 {
		t.Fatalf("expected non-empty response body")
	}
}
