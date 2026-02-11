# Releasing Sasswall

This repository uses GitHub Actions to build release artifacts and publish GitHub Releases when a version tag is pushed.

## Release workflow

Trigger: push a tag matching `v*`.

The workflow builds and uploads:

- `sasswall_linux_amd64`
- `sasswall_linux_arm64`
- `sha256sums.txt`

These names are intentionally stable because the auto-updater expects them.

## Procedure

1. Update `CHANGELOG.md` (move items from `[Unreleased]` into the new version section).
2. Commit changes.
3. Create and push a tag:

```bash
git tag v0.1.1
git push origin v0.1.1
```

4. Confirm the GitHub Release was created and contains the expected assets.

## Version metadata

The release workflow sets build metadata in the binary:

- `main.version` = tag name (e.g., `v0.1.1`)
- `main.commit` = short commit SHA

Validate locally:

```bash
./sasswall -version
```
