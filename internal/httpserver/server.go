package httpserver

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"sasswall/internal/canary"
	"sasswall/internal/classify"
	"sasswall/internal/config"
	"sasswall/internal/deception"
	"sasswall/internal/escalation"
	"sasswall/internal/metrics"
	"sasswall/internal/persona"
	"sasswall/internal/ratelimit"
	"sasswall/internal/session"
	"sasswall/internal/tarpit"
	"sasswall/internal/telemetry"
)

type Server struct {
	mu      sync.RWMutex
	cfg     *config.Config
	logger  *telemetry.Logger
	metrics *metrics.Registry

	classifier *classify.Classifier
	persona    *persona.Selector
	deception  *deception.Engine
	tarpit     *tarpit.Calculator
	strikes    *escalation.Tracker
	limiter    *ratelimit.Limiter
	sessions   *session.Store
	canary     *canary.Service

	inflight chan struct{}
	httpSrv  *http.Server
	metSrv   *http.Server
}

func New(cfg *config.Config, l *telemetry.Logger, m *metrics.Registry) (*Server, error) {
	c, err := classify.New(cfg.Honey.Paths, cfg.Fingerprints.LibraryFile)
	if err != nil {
		return nil, err
	}
	s := &Server{
		cfg:        cfg,
		logger:     l,
		metrics:    m,
		classifier: c,
		persona:    persona.New(cfg.Persona.Default, cfg.Persona.Dynamic, cfg.Persona.Mapping),
		deception:  deception.NewDefault(),
		tarpit: tarpit.New(tarpit.Config{
			DelayMin:    cfg.Tarpit.DelayMin,
			DelayMax:    cfg.Tarpit.DelayMax,
			HoneyBoost:  cfg.Honey.Boost,
			StrikeBoost: cfg.Escalation.StrikeBoost,
		}),
		strikes: escalation.New(cfg.Escalation.StrikeWindow, cfg.Escalation.StrikeThreshold, cfg.Escalation.DenyFor),
		limiter: ratelimit.New(ratelimit.Config{
			Burst:        cfg.RateLimit.Burst,
			NormalRPM:    cfg.RateLimit.NormalRPM,
			HoneyRPM:     cfg.RateLimit.HoneyRPM,
			ScannerRPM:   cfg.RateLimit.ScannerRPM,
			DeniedRPM:    cfg.RateLimit.DeniedRPM,
			Adaptive:     cfg.AdaptiveRate.Enabled,
			DegradeMax:   cfg.AdaptiveRate.DegradeSteps,
			RecoverAfter: cfg.AdaptiveRate.RecoverAfter,
		}),
		sessions: session.NewStore(cfg.SessionNarrative.TTL),
		inflight: make(chan struct{}, cfg.MaxInflight),
	}
	if cfg.Canary.Enabled {
		s.canary = canary.New(cfg.Canary.Secret)
	}
	s.httpSrv = &http.Server{Addr: cfg.Listen, Handler: http.HandlerFunc(s.handle)}
	if cfg.Metrics.Enabled {
		s.metSrv = &http.Server{Addr: cfg.Metrics.Listen, Handler: m.Handler()}
	}
	return s, nil
}

func (s *Server) Reload(cfg *config.Config) error {
	c, err := classify.New(cfg.Honey.Paths, cfg.Fingerprints.LibraryFile)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
	s.classifier = c
	s.persona = persona.New(cfg.Persona.Default, cfg.Persona.Dynamic, cfg.Persona.Mapping)
	s.tarpit = tarpit.New(tarpit.Config{DelayMin: cfg.Tarpit.DelayMin, DelayMax: cfg.Tarpit.DelayMax, HoneyBoost: cfg.Honey.Boost, StrikeBoost: cfg.Escalation.StrikeBoost})
	if cfg.Canary.Enabled {
		s.canary = canary.New(cfg.Canary.Secret)
	} else {
		s.canary = nil
	}
	return nil
}

func (s *Server) Start(ctx context.Context) error {
	errC := make(chan error, 2)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- err
		}
	}()
	if s.metSrv != nil {
		go func() {
			if err := s.metSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errC <- err
			}
		}()
	}
	select {
	case <-ctx.Done():
		_ = s.httpSrv.Shutdown(context.Background())
		if s.metSrv != nil {
			_ = s.metSrv.Shutdown(context.Background())
		}
		return nil
	case err := <-errC:
		return err
	}
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if !s.acquire() {
		http.Error(w, "service saturated", http.StatusServiceUnavailable)
		s.metrics.Inc("sasswall_requests_total", map[string]string{"status": "503", "category": "saturated"})
		return
	}
	defer s.release()

	s.mu.RLock()
	cfg := s.cfg
	classifier := s.classifier
	prs := s.persona
	dec := s.deception
	trp := s.tarpit
	cnr := s.canary
	s.mu.RUnlock()

	now := time.Now()
	ip := clientIP(r)
	ua := r.UserAgent()
	sessionID := session.SessionKey(ip, ua, now)
	_, denied, denyUntil := s.strikes.State(ip, now)
	cResult := classifier.Classify(r, denied)
	story := s.sessions.Update(sessionID, cResult.ScoreDelta, now)

	// hostile categories increase strike pressure
	if cResult.Category == classify.CategoryHoney || cResult.Category == classify.CategoryScanner {
		strikes, isDenied, until := s.strikes.AddStrike(ip, now)
		if isDenied {
			cResult.Category = classify.CategoryDenied
			denied = true
			denyUntil = until
			story = s.sessions.Update(sessionID, 3, now)
			_ = strikes
		}
	}

	rlDecision := s.limiter.Allow(ip, cResult.Category, story.Score, now)
	if !rlDecision.Allowed {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", rlDecision.RetryAfter))
		w.Header().Set("X-Sasswall-Category", string(cResult.Category))
		http.Error(w, "rate limited", http.StatusTooManyRequests)
		s.emitLog(start, telemetry.Event{
			IP: ip, UA: ua, Host: r.Host, Path: r.URL.Path,
			Category: string(cResult.Category), Persona: prs.Pick(cResult.Category),
			Denied: denied, DenyUntil: denyUntil.Format(time.RFC3339), Limited: true,
			SessionID: story.SessionID, SequenceScore: story.Score,
			FairnessStep: rlDecision.Step,
		})
		s.metrics.Inc("sasswall_requests_total", map[string]string{"status": "429", "category": string(cResult.Category)})
		return
	}

	pack := dec.Pick(story.SessionID, cfg.Deception.SurfacePacks.RotateEvery, now)
	profileID := pack.ID
	deceptionVariant := "none"
	decoySuccess := false
	canaryTokenID := ""

	if cfg.Deception.ReconPoison.Enabled {
		for k, v := range pack.Headers {
			w.Header().Set(k, v)
		}
		for k, v := range cfg.Deception.ReconPoison.HeaderSet {
			w.Header().Set(k, v)
		}
	}

	challengeIssued := false
	if cfg.Challenge.Enabled && categoryAllowed(cResult.Category, cfg.Challenge.AllowedCategories) {
		ok := false
		switch cfg.Challenge.Mode {
		case "cookie302":
			if c, err := r.Cookie("sw_challenge"); err == nil && c.Value == "ok" {
				ok = true
			}
			if !ok {
				challengeIssued = true
				http.SetCookie(w, &http.Cookie{Name: "sw_challenge", Value: "ok", MaxAge: 120, HttpOnly: true, Path: "/"})
				http.Redirect(w, r, r.URL.String(), http.StatusFound)
				s.metrics.Inc("sasswall_challenge_total", map[string]string{"mode": cfg.Challenge.Mode, "result": "issued"})
				s.emitLog(start, telemetry.Event{IP: ip, UA: ua, Host: r.Host, Path: r.URL.Path, Category: string(cResult.Category), Persona: prs.Pick(cResult.Category), SessionID: story.SessionID, SequenceScore: story.Score, ProfileID: profileID, ChallengeIssued: challengeIssued, FairnessStep: rlDecision.Step})
				return
			}
		case "pow-lite":
			if r.Header.Get("X-Sasswall-Pow") == "42" {
				ok = true
			}
			if !ok {
				challengeIssued = true
				w.Header().Set("X-Sasswall-Challenge", "pow-lite")
				http.Error(w, "challenge required", http.StatusTooManyRequests)
				s.metrics.Inc("sasswall_challenge_total", map[string]string{"mode": cfg.Challenge.Mode, "result": "issued"})
				s.emitLog(start, telemetry.Event{IP: ip, UA: ua, Host: r.Host, Path: r.URL.Path, Category: string(cResult.Category), Persona: prs.Pick(cResult.Category), SessionID: story.SessionID, SequenceScore: story.Score, ProfileID: profileID, ChallengeIssued: challengeIssued, Limited: true, FairnessStep: rlDecision.Step})
				return
			}
		}
		_ = ok
	}

	body := pack.Body
	status := http.StatusNotFound
	if cResult.Category == classify.CategoryDenied {
		status = http.StatusForbidden
	}

	if cfg.Deception.FakeBreadcrumbs.Enabled {
		if b, ok := dec.Breadcrumb(r.URL.Path, pack); ok {
			body = b
			deceptionVariant = "breadcrumb"
			decoySuccess = true
		}
	}

	if cfg.Canary.Enabled && cnr != nil {
		for _, tmpl := range cfg.Canary.HoneyFileTemplates {
			if strings.Contains(body, "__TOKEN__") {
				tok := cnr.Token(story.SessionID, r.URL.Path, story.Requests)
				body = canary.Inject(tmpl, tok)
				canaryTokenID = tok
				deceptionVariant = "canary"
				decoySuccess = true
				break
			}
		}
	}

	if shouldDelayedSuccess(cfg, cResult.Category, r.URL.Path, story) {
		status = http.StatusOK
		deceptionVariant = "delayed_success"
		decoySuccess = true
		s.sessions.MarkDelayedSuccess(story.SessionID)
		s.metrics.Inc("sasswall_delayed_success_total", map[string]string{"category": string(cResult.Category)})
	}

	delay := trp.Delay(story.Score, cResult.Honey, story.Score/3)
	time.Sleep(delay)

	w.Header().Set("X-Sasswall-Category", string(cResult.Category))
	w.Header().Set("X-Sasswall-Profile", profileID)
	w.Header().Set("X-Sasswall-Narrative", string(story.Phase))
	if status == http.StatusTooManyRequests {
		w.Header().Set("Retry-After", "2")
	}
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))

	s.metrics.Inc("sasswall_requests_total", map[string]string{"status": fmt.Sprintf("%d", status), "category": string(cResult.Category)})
	s.metrics.Observe("sasswall_tarpit_seconds", map[string]string{"category": string(cResult.Category)}, delay)
	s.emitLog(start, telemetry.Event{
		IP: ip, UA: ua, Host: r.Host, Path: r.URL.Path,
		Category: string(cResult.Category), Persona: prs.Pick(cResult.Category),
		Denied: denied, DenyUntil: denyUntil.Format(time.RFC3339), Limited: false,
		SessionID: story.SessionID, SequenceScore: story.Score,
		ProfileID: profileID, DeceptionVariant: deceptionVariant,
		ChallengeIssued: challengeIssued, DecoySuccess: decoySuccess,
		CanaryTokenID: canaryTokenID, FairnessStep: rlDecision.Step,
	})
}

func (s *Server) emitLog(start time.Time, e telemetry.Event) {
	e.Timestamp = time.Now().UTC()
	e.LatencyMS = time.Since(start).Milliseconds()
	s.logger.Emit(e)
}

func shouldDelayedSuccess(cfg *config.Config, category classify.Category, path string, story *session.Story) bool {
	if !cfg.Deception.DelayedSuccess.Enabled {
		return false
	}
	if category == classify.CategoryNormal {
		return false
	}
	if !categoryAllowed(category, []string{"scanner", "honey", "denied"}) {
		return false
	}
	if len(cfg.Deception.DelayedSuccess.AllowedPaths) > 0 && !pathAllowed(path, cfg.Deception.DelayedSuccess.AllowedPaths) {
		return false
	}
	if story.Requests < 3 {
		return false
	}
	ratio := float64(story.DelayedSuccess) / float64(max(story.Requests, 1))
	if ratio >= cfg.Deception.DelayedSuccess.MaxRatio {
		return false
	}
	return rand.Intn(100) < int(cfg.Deception.DelayedSuccess.MaxRatio*100)
}

func pathAllowed(path string, patterns []string) bool {
	p := strings.ToLower(path)
	for _, raw := range patterns {
		v := strings.ToLower(raw)
		switch {
		case strings.Contains(v, "*"):
			s := strings.ReplaceAll(v, "*", "")
			if s != "" && strings.Contains(p, s) {
				return true
			}
		case strings.HasSuffix(v, "/"):
			if strings.HasPrefix(p, v) {
				return true
			}
		default:
			if p == v {
				return true
			}
		}
	}
	return false
}

func categoryAllowed(cat classify.Category, allow []string) bool {
	if len(allow) == 0 {
		return false
	}
	for _, v := range allow {
		if strings.EqualFold(v, string(cat)) {
			return true
		}
	}
	return false
}

func clientIP(r *http.Request) string {
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}

func (s *Server) acquire() bool {
	select {
	case s.inflight <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Server) release() {
	select {
	case <-s.inflight:
	default:
		_, _ = fmt.Fprintln(os.Stderr, "inflight underflow")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
