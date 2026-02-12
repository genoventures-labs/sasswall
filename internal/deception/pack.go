package deception

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

type Pack struct {
	ID      string
	Headers map[string]string
	Body    string
	Decoys  map[string]string
}

type Engine struct {
	packs []Pack
}

func NewDefault() *Engine {
	return &Engine{packs: []Pack{
		{
			ID:      "php-admin",
			Headers: map[string]string{"Server": "Apache", "X-Powered-By": "PHP/7.4", "Cache-Control": "private, max-age=0"},
			Body:    "<html><body><h1>Admin Gateway</h1><p>Sign in required.</p></body></html>",
			Decoys: map[string]string{
				"/.env":        "APP_ENV=production\nDB_HOST=127.0.0.1\nDB_PASSWORD=__TOKEN__\n",
				"/.git/config": "[core]\nrepositoryformatversion = 0\nfilemode = true\n",
			},
		},
		{
			ID:      "wp-old",
			Headers: map[string]string{"Server": "nginx", "X-Powered-By": "PHP/5.6", "Cache-Control": "no-store"},
			Body:    "<html><body><h1>WordPress Maintenance</h1><p>Temporarily unavailable.</p></body></html>",
			Decoys: map[string]string{
				"/wp-login.php": "<html><body><form><input name='log'/><input name='pwd'/></form></body></html>",
				"/phpmyadmin":   "<html><body><h1>phpMyAdmin</h1><p>Session expired.</p></body></html>",
			},
		},
		{
			ID:      "generic-enterprise",
			Headers: map[string]string{"Server": "edge-gateway", "X-Powered-By": "ASP.NET", "Cache-Control": "private, no-cache"},
			Body:    "<html><body><h1>Portal</h1><p>Access check in progress.</p></body></html>",
			Decoys: map[string]string{
				"/admin": "<html><body><h1>Admin Portal</h1><p>Unauthorized</p></body></html>",
			},
		},
	}}
}

func (e *Engine) Pick(sessionID string, rotateEvery time.Duration, now time.Time) Pack {
	if len(e.packs) == 0 {
		return Pack{ID: "none"}
	}
	if rotateEvery <= 0 {
		rotateEvery = 15 * time.Minute
	}
	bucket := now.Unix() / int64(rotateEvery.Seconds())
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", sessionID, bucket)))
	idx := int(h[0]) % len(e.packs)
	return e.packs[idx]
}

func (e *Engine) Breadcrumb(path string, pack Pack) (string, bool) {
	for k, v := range pack.Decoys {
		if matchPath(path, k) {
			return v, true
		}
	}
	return "", false
}

func matchPath(path, pattern string) bool {
	p := strings.ToLower(pattern)
	v := strings.ToLower(path)
	switch {
	case strings.Contains(p, "*"):
		s := strings.ReplaceAll(p, "*", "")
		return s != "" && strings.Contains(v, s)
	case strings.HasSuffix(p, "/"):
		return strings.HasPrefix(v, p)
	default:
		return v == p
	}
}
