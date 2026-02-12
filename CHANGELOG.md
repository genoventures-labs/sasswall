# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- 

### Changed
- 

### Fixed
- 

## [0.2.0] - 2026-02-12

### Added
- Greenfield Sasswall runtime implementation (`cmd/sasswall`) with YAML config loading, SIGHUP reload path, and version metadata output.
- Threat Theater engine with deterministic deceptive surface packs (`php-admin`, `wp-old`, `generic-enterprise`).
- Session narratives (`observe -> engage -> sink`) and sequence scoring.
- Built-in scanner fingerprint library with optional external fingerprint file support.
- Fake breadcrumb responses tied to deception profile selection.
- Canary token generation and offline verification command (`sasswall verify-canary`).
- Adaptive fairness-aware rate limiting with degradation steps.
- Prometheus-style metrics endpoint on dedicated listener (`/metrics`).
- Structured telemetry fields: `session_id`, `sequence_score`, `profile_id`, `deception_variant`, `challenge_issued`, `decoy_success`, `canary_token_id`, and `fairness_step`.
- Threat Theater config schema additions for deception, canary, session narrative, fingerprints, adaptive rate, challenge, and metrics.
- Initial unit/integration tests for config validation, classifier behavior, tarpit cap, canary determinism, and HTTP handler response headers.

### Changed
- README, operations, and security docs expanded for Threat Theater behavior, safety boundaries, metrics, and canary operations.
- Partial-default rollout enforced: low-risk deception toggles enabled by default, higher-risk capabilities opt-in.

### Fixed
- N/A

## [0.1.1] - 2026-02-11

### Changed
- Updated README badges/links to point at the published GitHub Releases and changelog.

## [0.1.0] - 2026-02-11

### Added
- SDSA core: tarpit delays (initial + per-character streaming) and phrase randomization.
- Honey endpoints and scanner UA detection.
- Adaptive escalation: per-IP strikes and temporary deny mode.
- Per-category rate limiting and load shedding.
- Persona system (sister/it/hr), with dynamic per-category selection.
- YAML configuration and SIGHUP hot reload.
- systemd service unit defaults.
