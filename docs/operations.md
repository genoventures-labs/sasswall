# Sasswall operations

## Service control

```bash
systemctl status sasswall --no-pager
systemctl restart sasswall
systemctl reload sasswall   # hot reload config via SIGHUP
```

## Logs

```bash
journalctl -u sasswall -f --no-pager
```

Key fields emitted per request:

- `ip`, `ua`, `host`, `path`
- `category`: `normal|scanner|honey|denied`
- `persona`: `sister|it|hr` (configurable)
- `strikes`, `denied`, `deny_until`
- `limited`: whether a request was rate limited
- `session_id`, `sequence_score`, `profile_id`
- `deception_variant`, `challenge_issued`, `decoy_success`
- `canary_token_id`, `fairness_step`
- `presence_state`, `presence_transition`, `presence_signature_id`, `pressure_action_applied`

## Metrics

When `metrics.enabled=true`, Sasswall exposes Prometheus text metrics on a dedicated listener:

```bash
curl -s http://127.0.0.1:9182/metrics
```

Core metrics include category/status counters and tarpit latency buckets.
Presence mode metrics include:

- `sasswall_presence_state_total{state}`
- `sasswall_presence_transition_total{from,to}`
- `sasswall_presence_pressure_action_total{action}`

## Canary verification (offline)

For canary-enabled deployments, verify a token offline:

```bash
sasswall verify-canary \
  --secret 'YOUR_SECRET' \
  --token 'TOKEN' \
  --session 'SESSION_ID' \
  --path '/.env' \
  --n 1
```

## Config changes

1. Edit `/etc/sasswall/config.yaml`.
2. Apply without restart:

```bash
systemctl reload sasswall
```

Notes:
- Most settings reload immediately.
- `listen` and `metrics.listen` require restart.

## Troubleshooting

### Confirm config is being used

Check the ExecStart and startup log:

```bash
systemctl show -p ExecStart sasswall
journalctl -u sasswall -n 20 --no-pager
```

### Validate YAML

Common failures:
- invalid YAML syntax
- missing quotes around `*...*` patterns

Sasswall logs config reload failures; service continues using the last known-good configuration.
