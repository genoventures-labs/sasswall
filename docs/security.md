# Security posture

## Scope

Sasswall is an SDSA component that provides application-layer friction, deception primitives (honey paths), and security telemetry.

It is not a replacement for:

- firewalling and segmentation
- WAF/IDS/IPS
- authentication/authorization
- patch management
- DDoS mitigation

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
