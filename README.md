# Sasswall — Sass-Driven Security Architecture (SDSA)

A lightweight, daemonized HTTP “noise sink” that absorbs hostile/scanner traffic, wastes attacker time, and generates high-signal telemetry—while keeping production applications quieter and harder to fingerprint.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Linux-blue)](#)
[![Service](https://img.shields.io/badge/systemd-supported-6c757d)](#)
[![Security](https://img.shields.io/badge/Security-Defense--in--Depth-important)](#)
[![Release](https://img.shields.io/github/v/release/cassianwolfe/sasswall?display_name=tag&sort=semver)](https://github.com/cassianwolfe/sasswall/releases)
[![Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-informational)](CHANGELOG.md)

---

## Table of Contents

- [Overview](#overview)
- [Why SDSA](#why-sdsa)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Hot reload](#hot-reload)
- [Reverse proxy integration (Caddy)](#reverse-proxy-integration-caddy)
- [Operations](#operations)
- [Security model](#security-model)

---

## Overview

**Sasswall** is an application-layer security component intended to sit behind your reverse proxy (Caddy/Nginx/Envoy) and handle traffic that should never reach production services:

- unknown hosts and by-IP requests
- common opportunistic probes (e.g., `/.env`, `/.git/`, `/wp-login.php`, `phpmyadmin`)
- scanner user-agents (e.g., nmap/masscan/zgrab/sqlmap/nikto)

Instead of leaking fingerprints, Sasswall returns consistent responses, intentionally introduces latency (“tarpit”), and emits clear logs for threat triage.

---

## Why SDSA

**Sass-Driven Security Architecture (SDSA)** is a defense-in-depth pattern designed to:

1. **Absorb and degrade** low-value hostile traffic (probes, opportunistic scans, bot noise).
2. **Reduce attacker signal** (no application fingerprinting; generic behavior).
3. **Increase attacker cost** (bounded latency, adaptive escalation, throttling).
4. **Create actionable telemetry** (high-signal logs you can forward to SIEM/WAF/firewall rules).

SDSA does not replace firewalls, WAFs, authentication, or patching. It complements them.

---

## Key Features

### Persona-driven responses (static or dynamic)
Sasswall supports response “personas” (enterprise-friendly wrappers) and can select them per request category:

- `normal` → `sister` (default)
- `scanner` → `it`
- `honey` → `it`
- `denied` → `hr`

Persona mapping is configurable.

### Honey endpoints
Recognizes common probe paths and treats them as honey endpoints:

- additional tarpit delay
- configurable stricter rate limits
- high-signal logs

### Adaptive escalation (strikes → temporary deny mode)
Maintains lightweight per-IP strike tracking within a time window. Above threshold, an IP enters a temporary deny mode.

### Tarpit mode (bounded delay + optional slow streaming)
Intentionally slows responses using:

- randomized initial delay
- per-character streaming delay
- additional delay for scanner-like user agents

### Rate limiting and load shedding
- per-IP token bucket rate limiting by category (`normal`, `scanner`, `honey`, `denied`)
- concurrency cap (`max_inflight`)
- returns `429` with `Retry-After` when limited
- returns `503` when saturated

### Threat Theater (defensive deception mode)
Sasswall v0.2.0 adds a configurable defensive deception layer:

- deterministic deceptive surface packs (`php-admin`, `wp-old`, `generic-enterprise`)
- session narratives (`observe -> engage -> sink`) with sequence-scored escalation
- recon-poison headers and fake breadcrumbs on hostile probes
- adaptive fairness rate step-down before hard deny
- optional canary artifact tokenization (offline verification only)
- optional delayed-success illusions (controlled fake `200` for hostile categories only)
- optional challenge gates (`cookie302` or `pow-lite`)

High-risk features are opt-in by default.

### Predator Presence mode (opt-in)
`presence` is an opt-in Threat Theater profile that creates a coherent hostile-session shift:

- state progression: `observe -> lock-on -> pressure`
- deterministic session signature over a coherence window
- pressure actions applied in configured order
- hostile-only response header: `X-Sasswall-Presence`

---

## Architecture

A common placement pattern:

```
Internet
  |
  |  (optional) CDN/WAF
  v
Reverse Proxy (Caddy/Nginx/Envoy)
  |-----------------------> Production Apps
  |
  +-----------------------> Sasswall (unknown hosts / :80 catch-all / honey paths)
```

See [`docs/architecture.md`](docs/architecture.md) for the detailed SDSA model.

---

## Installation

Sasswall is typically deployed as:

- Binary: `/usr/local/bin/sasswall`
- Config: `/etc/sasswall/config.yaml`
- Phrase list: `/etc/sasswall/phrases.txt`
- Systemd unit: `/etc/systemd/system/sasswall.service`

Start/enable:

```bash
systemctl daemon-reload
systemctl enable --now sasswall
systemctl status sasswall --no-pager
```

Smoke test:

```bash
curl -i http://127.0.0.1:8182/
curl -i http://127.0.0.1:8182/.env
curl -i -A 'nmap' http://127.0.0.1:8182/
```

---

## Configuration

Sasswall runs from a single YAML config:

- Default path: `/etc/sasswall/config.yaml`
- Start flag: `sasswall -config /etc/sasswall/config.yaml`

Example configuration is included below; the same schema is used in production.

```yaml
listen: 127.0.0.1:8182
phrases_file: /etc/sasswall/phrases.txt
max_inflight: 128

persona:
  default: sister
  dynamic: true
  mapping:
    normal: sister
    scanner: it
    honey: it
    denied: hr

tarpit:
  delay_min: 800ms
  delay_max: 2500ms
  char_min: 15ms
  char_max: 40ms

honey:
  enabled: true
  boost: 2s
  paths:
    - /.env
    - /.git/config
    - /.git/
    - /wp-login.php
    - /wp-admin/
    - /phpmyadmin
    - "*phpmyadmin*"
    - "*.env*"

escalation:
  strike_window: 10m
  strike_threshold: 6
  deny_for: 30m
  strike_boost: 200ms

rate_limit:
  enabled: true
  burst: 12
  normal_rpm: 120
  honey_rpm: 40
  scanner_rpm: 60
  denied_rpm: 15

threat_theater:
  enabled: false
  profile: balanced

presence:
  enabled: false
  signal_style: subtle
  lock_on_threshold: 4
  pressure_threshold: 8
  coherence_window: 30m
  header_signature:
    enabled: true
    rotation: 15m
  timing_signature:
    enabled: true
    jitter_band_ms: 120
  pressure_actions:
    - tarpit_boost
    - fairness_stepup
    - challenge_hint

deception:
  surface_packs:
    enabled: true
    rotate_every: 15m
    allowed_categories: [scanner, honey, denied]
  fake_breadcrumbs:
    enabled: true
  recon_poison_headers:
    enabled: true
    header_set:
      Server: edge-gateway
      X-Powered-By: PHP/7.4
  delayed_success:
    enabled: false
    max_ratio: 0.10
    allowed_paths:
      - /.env
      - /.git/config
      - /wp-login.php
      - /phpmyadmin
      - "*phpmyadmin*"
      - "*.env*"

session_narrative:
  enabled: true
  ttl: 30m
  style: subtle

fingerprints:
  enabled: true
  library_file: ""

adaptive_rate:
  enabled: true
  degrade_steps: 3
  recover_after: 20m

challenge:
  enabled: false
  mode: cookie302
  allowed_categories: [scanner, honey]

canary:
  enabled: false
  secret: ""
  honey_file_templates:
    - "DB_PASSWORD=__TOKEN__"
    - "AWS_SECRET_ACCESS_KEY=__TOKEN__"

metrics:
  enabled: true
  listen: 127.0.0.1:9182
```

Honey path matching rules:

- Exact match: `/.env`
- Prefix match (ends with `/`): `/.git/`
- Substring match (contains `*`): `*phpmyadmin*`

---

## Hot reload

Sasswall supports hot reload via `SIGHUP`.

With systemd:

```bash
systemctl reload sasswall
```

Reloads without restart:

- personas and category mapping
- honey paths
- tarpit timings
- escalation and rate limits
- fingerprints library file
- threat theater/deception/challenge/canary behavior

Requires restart:

- `listen` (bind address)

---

## Auto-updating (GitHub Releases)

This host can keep Sasswall up to date automatically using a systemd timer:

- `sasswall-update.timer` (scheduled update checks)
- `sasswall-update.service` (one-shot update run)

The updater checks **GitHub Releases**, downloads the correct Linux binary for the host architecture, verifies checksums (recommended), installs to `/usr/local/bin/sasswall`, and restarts the `sasswall` service.

Documentation:
- [`docs/auto-update.md`](docs/auto-update.md)
- [`docs/releasing.md`](docs/releasing.md)
- [`CHANGELOG.md`](CHANGELOG.md)

---

## Reverse proxy integration (Caddy)

Sasswall is designed to be reverse-proxied and typically bound to `127.0.0.1`.

- Use it as a `:80` catch-all sink
- Optionally route specific honey paths to it (if you want to centralize scanning noise)

See [`docs/deployment-caddy.md`](docs/deployment-caddy.md) for copy/paste patterns.

---

## Operations

### Logs

```bash
journalctl -u sasswall -f --no-pager
```

Log entries include:

- `ip`, `ua`, `path`, `host`
- `category` (`normal|scanner|honey|denied`)
- `persona`
- `strikes`, `denied`, `deny_until`
- `limited`
- `session_id`, `sequence_score`, `profile_id`
- `deception_variant`, `challenge_issued`, `decoy_success`
- `canary_token_id`, `fairness_step`
- `presence_state`, `presence_transition`, `presence_signature_id`, `pressure_action_applied`

Responses include:

- `X-Sasswall-Category: <category>`
- `X-Sasswall-Profile: <profile-id>`
- `X-Sasswall-Narrative: <observe|engage|sink>`
- `X-Sasswall-Presence: <observe|lock-on|pressure>` (hostile categories, presence mode only)

Metrics endpoint (`Prometheus` text format):

```bash
curl -s http://127.0.0.1:9182/metrics
```

---

## Security model

Sasswall provides **application-layer friction and telemetry**. It is not a perimeter firewall.

Recommended defense-in-depth:

- firewall/ACLs restricting origin access
- strict reverse-proxy host routing (no default proxy-to-app)
- optional CDN/WAF
- log forwarding to SIEM and alerting pipelines

Legal/safety boundary:

- Sasswall performs **defensive inbound deception only**.
- No exploitation, no outbound attacks, no interaction with attacker infrastructure.
- Canary tokens are generated and logged locally; exfiltration detection is expected to happen in external SIEM/workflows.
