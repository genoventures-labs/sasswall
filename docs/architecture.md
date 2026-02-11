# Sass-Driven Security Architecture (SDSA)

## Summary

Sass-Driven Security Architecture (SDSA) is a defensive pattern where **unwanted internet traffic is intentionally handled by a dedicated sink** that:

- reduces fingerprinting of production services
- increases attacker cost with bounded delay (tarpit)
- emits clean, high-signal telemetry
- supports deception (honey endpoints) and adaptive behavior

Sasswall is an SDSA component designed for deployment behind a reverse proxy.

---

## Threat model focus

SDSA targets the long tail of:

- opportunistic probing (`/.env`, `/.git/`, `phpmyadmin`, common CMS/admin paths)
- automated scanners and recon tools
- low-effort credential stuffing and endpoint discovery
- noisy by-IP or unknown host traffic

SDSA is not a replacement for:

- patch management
- authentication/authorization
- firewalling and segmentation
- WAF/IDS/IPS
- DDoS mitigation

---

## Control plane vs data plane

SDSA separates concerns:

- **Data plane:** Sasswall handles HTTP requests that should not reach apps.
- **Control plane:** your reverse proxy and perimeter controls decide which traffic is routed to Sasswall vs apps.

This separation keeps Sasswall simple and predictable.

---

## SDSA data flow

Typical flow:

1. Request arrives at reverse proxy.
2. Proxy evaluates routing criteria:
   - known hostnames → app
   - unknown/invalid hostnames → Sasswall
   - optionally: certain paths/headers → Sasswall
3. Sasswall categorizes the request:
   - `normal`, `scanner`, `honey`, `denied`
4. Sasswall applies controls:
   - rate limits
   - strike tracking and deny window
   - bounded tarpit latency
   - persona selection (static or per-category)
5. Sasswall emits telemetry:
   - `category`, `persona`, `strikes`, `limited`

---

## SDSA outcomes

### Reduced attack surface visibility
- No direct application fingerprints for unknown traffic.
- Consistent, policy-driven behavior instead of accidental leaks.

### Increased attacker cost
- Controlled time waste on probes.
- Adaptive behavior for repeat offenders.

### Better signals
- Cleaner logs for security analytics.
- Easy enrichment and forwarding to SIEM.

---

## Recommended deployment invariants

- Bind Sasswall to localhost (or private network only).
- Enforce strict host routing at the reverse proxy.
- Keep response content non-sensitive and uniform.
- Treat SDSA logs as a security telemetry stream.
