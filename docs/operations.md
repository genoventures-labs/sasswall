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

## Config changes

1. Edit `/etc/sasswall/config.yaml`.
2. Apply without restart:

```bash
systemctl reload sasswall
```

Notes:
- Most settings reload immediately.
- `listen` requires restart.

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
