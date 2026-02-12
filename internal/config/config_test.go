package config

import "testing"

func TestDefaultPartialRollout(t *testing.T) {
	cfg := Default()
	if !cfg.Deception.SurfacePacks.Enabled || !cfg.Deception.FakeBreadcrumbs.Enabled || !cfg.Deception.ReconPoison.Enabled {
		t.Fatalf("expected low-risk deception defaults enabled")
	}
	if cfg.Deception.DelayedSuccess.Enabled || cfg.Canary.Enabled || cfg.Challenge.Enabled || cfg.ThreatTheater.Enabled {
		t.Fatalf("expected high-risk features disabled by default")
	}
}

func TestValidateCanarySecretRequired(t *testing.T) {
	cfg := Default()
	cfg.Canary.Enabled = true
	cfg.Canary.Secret = ""
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for missing canary secret")
	}
}

func TestValidateRatioBounds(t *testing.T) {
	cfg := Default()
	cfg.Deception.DelayedSuccess.MaxRatio = 2
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected ratio validation error")
	}
}

func TestValidatePresenceProfile(t *testing.T) {
	cfg := Default()
	cfg.ThreatTheater.Profile = "presence"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected presence profile to validate: %v", err)
	}
}

func TestValidatePresenceSignalStyle(t *testing.T) {
	cfg := Default()
	cfg.Presence.SignalStyle = "bad"
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected presence.signal_style validation error")
	}
}
