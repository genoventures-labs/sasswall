# Security posture

## Scope

Sasswall is an SDSA component that provides application-layer friction, deception primitives (honey paths), and security telemetry.

It is not a replacement for:

- firewalling and segmentation
- WAF/IDS/IPS
- authentication/authorization
- patch management
- DDoS mitigation

## Defensive deception boundary

Threat Theater features are strictly defensive and inbound:

- no exploitation attempts
- no outbound attacks
- no interaction with attacker infrastructure

Deception mechanisms (surface packs, breadcrumbs, delayed-success illusions, and challenge gates) are designed to increase attacker uncertainty and cost while preserving legal/operational safety.

## Update integrity

The host-side updater can be configured to require checksums (`SASSWALL_REQUIRE_CHECKSUMS=true`).

For enterprise deployments, prefer:

- enforcing checksum verification
- private repos with a least-privilege PAT for updates
- auditable release pipelines (GitHub Actions with protected branches/tags)

## Logging

Logs are intended as a security telemetry stream. Treat them accordingly:

- forward to a central log store
- apply retention policies
- consider enrichment (GeoIP, ASN) upstream

Canary tokens are generated and logged locally. Detection and response for token reuse should be handled by external SIEM/monitoring workflows.
