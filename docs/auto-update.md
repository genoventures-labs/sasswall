# Auto-updating Sasswall (GitHub Releases)

Sasswall supports host-side auto-updates using systemd.

## Components

- `/usr/local/sbin/sasswall-update` — updater script
- `/etc/sasswall/updater.env` — updater configuration (repo + token)
- `sasswall-update.service` — one-shot update unit
- `sasswall-update.timer` — scheduled update checks

## Setup

1. Edit `/etc/sasswall/updater.env`:

- Set `SASSWALL_REPO=OWNER/REPO`
- For **private** repositories, set `GITHUB_TOKEN=...` (PAT with `repo` scope)

2. Enable the timer:

```bash
systemctl daemon-reload
systemctl enable --now sasswall-update.timer
```

3. Run an immediate update check:

```bash
systemctl start sasswall-update.service
```

## Logs

```bash
journalctl -u sasswall-update.service -n 200 --no-pager
```

## Supply-chain expectations (recommended)

For enterprise deployments, keep `SASSWALL_REQUIRE_CHECKSUMS=true` and ensure each GitHub Release includes:

- `sasswall_linux_amd64`
- `sasswall_linux_arm64`
- `sha256sums.txt` (must include hash lines for the above asset names)

If checksums are missing and `SASSWALL_REQUIRE_CHECKSUMS=true`, the updater refuses to install.

## Manual update

You can always run the updater directly with the environment file:

```bash
set -a
source /etc/sasswall/updater.env
set +a
/usr/local/sbin/sasswall-update
```
