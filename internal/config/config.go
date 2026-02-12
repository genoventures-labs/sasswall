package config

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	ErrCanarySecretRequired = errors.New("canary.secret is required when canary.enabled=true")
)

type Config struct {
	Listen      string        `yaml:"listen"`
	PhrasesFile string        `yaml:"phrases_file"`
	MaxInflight int           `yaml:"max_inflight"`
	Persona     PersonaConfig `yaml:"persona"`
	Tarpit      TarpitConfig  `yaml:"tarpit"`
	Honey       HoneyConfig   `yaml:"honey"`
	Escalation  Escalation    `yaml:"escalation"`
	RateLimit   RateLimit     `yaml:"rate_limit"`

	ThreatTheater    ThreatTheater    `yaml:"threat_theater"`
	Deception        Deception        `yaml:"deception"`
	Canary           Canary           `yaml:"canary"`
	SessionNarrative SessionNarrative `yaml:"session_narrative"`
	Fingerprints     Fingerprints     `yaml:"fingerprints"`
	AdaptiveRate     AdaptiveRate     `yaml:"adaptive_rate"`
	Challenge        Challenge        `yaml:"challenge"`
	Metrics          Metrics          `yaml:"metrics"`
}

type PersonaConfig struct {
	Default string            `yaml:"default"`
	Dynamic bool              `yaml:"dynamic"`
	Mapping map[string]string `yaml:"mapping"`
}

type TarpitConfig struct {
	DelayMin time.Duration `yaml:"delay_min"`
	DelayMax time.Duration `yaml:"delay_max"`
	CharMin  time.Duration `yaml:"char_min"`
	CharMax  time.Duration `yaml:"char_max"`
}

type HoneyConfig struct {
	Enabled bool          `yaml:"enabled"`
	Boost   time.Duration `yaml:"boost"`
	Paths   []string      `yaml:"paths"`
}

type Escalation struct {
	StrikeWindow    time.Duration `yaml:"strike_window"`
	StrikeThreshold int           `yaml:"strike_threshold"`
	DenyFor         time.Duration `yaml:"deny_for"`
	StrikeBoost     time.Duration `yaml:"strike_boost"`
}

type RateLimit struct {
	Enabled    bool `yaml:"enabled"`
	Burst      int  `yaml:"burst"`
	NormalRPM  int  `yaml:"normal_rpm"`
	HoneyRPM   int  `yaml:"honey_rpm"`
	ScannerRPM int  `yaml:"scanner_rpm"`
	DeniedRPM  int  `yaml:"denied_rpm"`
}

type ThreatTheater struct {
	Enabled bool   `yaml:"enabled"`
	Profile string `yaml:"profile"`
}

type Deception struct {
	SurfacePacks    SurfacePacks   `yaml:"surface_packs"`
	FakeBreadcrumbs FeatureToggle  `yaml:"fake_breadcrumbs"`
	ReconPoison     ReconPoison    `yaml:"recon_poison_headers"`
	DelayedSuccess  DelayedSuccess `yaml:"delayed_success"`
}

type SurfacePacks struct {
	Enabled           bool          `yaml:"enabled"`
	RotateEvery       time.Duration `yaml:"rotate_every"`
	AllowedCategories []string      `yaml:"allowed_categories"`
}

type FeatureToggle struct {
	Enabled bool `yaml:"enabled"`
}

type ReconPoison struct {
	Enabled   bool              `yaml:"enabled"`
	HeaderSet map[string]string `yaml:"header_set"`
}

type DelayedSuccess struct {
	Enabled      bool     `yaml:"enabled"`
	MaxRatio     float64  `yaml:"max_ratio"`
	AllowedPaths []string `yaml:"allowed_paths"`
}

type Canary struct {
	Enabled            bool     `yaml:"enabled"`
	Secret             string   `yaml:"secret"`
	HoneyFileTemplates []string `yaml:"honey_file_templates"`
}

type SessionNarrative struct {
	Enabled bool          `yaml:"enabled"`
	TTL     time.Duration `yaml:"ttl"`
	Style   string        `yaml:"style"`
}

type Fingerprints struct {
	Enabled     bool   `yaml:"enabled"`
	LibraryFile string `yaml:"library_file"`
}

type AdaptiveRate struct {
	Enabled      bool          `yaml:"enabled"`
	DegradeSteps int           `yaml:"degrade_steps"`
	RecoverAfter time.Duration `yaml:"recover_after"`
}

type Challenge struct {
	Enabled           bool     `yaml:"enabled"`
	Mode              string   `yaml:"mode"`
	AllowedCategories []string `yaml:"allowed_categories"`
}

type Metrics struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

func Default() Config {
	return Config{
		Listen:      "127.0.0.1:8182",
		PhrasesFile: "/etc/sasswall/phrases.txt",
		MaxInflight: 128,
		Persona: PersonaConfig{
			Default: "sister",
			Dynamic: true,
			Mapping: map[string]string{
				"normal":  "sister",
				"scanner": "it",
				"honey":   "it",
				"denied":  "hr",
			},
		},
		Tarpit: TarpitConfig{
			DelayMin: 800 * time.Millisecond,
			DelayMax: 2500 * time.Millisecond,
			CharMin:  15 * time.Millisecond,
			CharMax:  40 * time.Millisecond,
		},
		Honey: HoneyConfig{
			Enabled: true,
			Boost:   2 * time.Second,
			Paths:   []string{"/.env", "/.git/config", "/.git/", "/wp-login.php", "/wp-admin/", "/phpmyadmin", "*phpmyadmin*", "*.env*"},
		},
		Escalation: Escalation{
			StrikeWindow:    10 * time.Minute,
			StrikeThreshold: 6,
			DenyFor:         30 * time.Minute,
			StrikeBoost:     200 * time.Millisecond,
		},
		RateLimit: RateLimit{
			Enabled:    true,
			Burst:      12,
			NormalRPM:  120,
			HoneyRPM:   40,
			ScannerRPM: 60,
			DeniedRPM:  15,
		},
		ThreatTheater: ThreatTheater{Enabled: false, Profile: "balanced"},
		Deception: Deception{
			SurfacePacks:    SurfacePacks{Enabled: true, RotateEvery: 15 * time.Minute, AllowedCategories: []string{"scanner", "honey", "denied"}},
			FakeBreadcrumbs: FeatureToggle{Enabled: true},
			ReconPoison: ReconPoison{Enabled: true, HeaderSet: map[string]string{
				"Server":                 "edge-gateway",
				"X-Powered-By":           "PHP/7.4",
				"Cache-Control":          "private, max-age=0, no-cache",
				"X-Frame-Options":        "SAMEORIGIN",
				"X-Content-Type-Options": "nosniff",
			}},
			DelayedSuccess: DelayedSuccess{Enabled: false, MaxRatio: 0.10, AllowedPaths: []string{"/.env", "/.git/config", "/wp-login.php", "/phpmyadmin", "*.env*", "*phpmyadmin*"}},
		},
		Canary:           Canary{Enabled: false, HoneyFileTemplates: []string{"DB_PASSWORD=__TOKEN__", "AWS_SECRET_ACCESS_KEY=__TOKEN__", "API_KEY=__TOKEN__"}},
		SessionNarrative: SessionNarrative{Enabled: true, TTL: 30 * time.Minute, Style: "subtle"},
		Fingerprints:     Fingerprints{Enabled: true},
		AdaptiveRate:     AdaptiveRate{Enabled: true, DegradeSteps: 3, RecoverAfter: 20 * time.Minute},
		Challenge:        Challenge{Enabled: false, Mode: "cookie302", AllowedCategories: []string{"scanner", "honey"}},
		Metrics:          Metrics{Enabled: true, Listen: "127.0.0.1:9182"},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyThreatTheaterOverlay()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Listen == "" {
		return errors.New("listen is required")
	}
	if c.MaxInflight <= 0 {
		return errors.New("max_inflight must be > 0")
	}
	if c.Tarpit.DelayMin <= 0 || c.Tarpit.DelayMax <= 0 || c.Tarpit.DelayMin > c.Tarpit.DelayMax {
		return errors.New("invalid tarpit delay range")
	}
	if c.Tarpit.CharMin <= 0 || c.Tarpit.CharMax <= 0 || c.Tarpit.CharMin > c.Tarpit.CharMax {
		return errors.New("invalid tarpit char range")
	}
	if c.Escalation.StrikeThreshold <= 0 {
		return errors.New("escalation.strike_threshold must be > 0")
	}
	if c.Deception.DelayedSuccess.MaxRatio < 0 || c.Deception.DelayedSuccess.MaxRatio > 1 {
		return errors.New("deception.delayed_success.max_ratio must be in [0,1]")
	}
	if c.Canary.Enabled && strings.TrimSpace(c.Canary.Secret) == "" {
		return ErrCanarySecretRequired
	}
	if !slices.Contains([]string{"conservative", "balanced", "aggressive"}, c.ThreatTheater.Profile) {
		return errors.New("threat_theater.profile must be conservative|balanced|aggressive")
	}
	if !slices.Contains([]string{"subtle", "theatrical"}, c.SessionNarrative.Style) {
		return errors.New("session_narrative.style must be subtle|theatrical")
	}
	if !slices.Contains([]string{"cookie302", "pow-lite"}, c.Challenge.Mode) {
		return errors.New("challenge.mode must be cookie302|pow-lite")
	}
	if c.AdaptiveRate.DegradeSteps < 1 {
		return errors.New("adaptive_rate.degrade_steps must be >= 1")
	}
	if c.Metrics.Enabled && c.Metrics.Listen == "" {
		return errors.New("metrics.listen is required when metrics.enabled=true")
	}
	return nil
}

func (c *Config) applyThreatTheaterOverlay() {
	if !c.ThreatTheater.Enabled {
		return
	}
	// Profile overlay intentionally only turns knobs up; user explicit toggles still win.
	switch c.ThreatTheater.Profile {
	case "conservative":
		if c.Deception.DelayedSuccess.Enabled {
			c.Deception.DelayedSuccess.Enabled = false
		}
		if c.Challenge.Enabled {
			c.Challenge.Enabled = false
		}
	case "aggressive":
		if c.Honey.Boost < 3*time.Second {
			c.Honey.Boost = 3 * time.Second
		}
		if c.Escalation.StrikeThreshold > 4 {
			c.Escalation.StrikeThreshold = 4
		}
		if c.Deception.DelayedSuccess.MaxRatio < 0.15 {
			c.Deception.DelayedSuccess.MaxRatio = 0.15
		}
	}
}
