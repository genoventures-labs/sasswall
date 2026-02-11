# Releasing Sasswall

Sasswall updates on hosts by consuming **GitHub Releases** that contain stable asset names.

## Required release assets

Each release should include:

- `sasswall_linux_amd64`
- `sasswall_linux_arm64`
- `sha256sums.txt`

These names are intentionally stable because the host-side auto-updater expects them.

## Option A (recommended): GitHub Actions workflow

A ready-to-use workflow template is provided at:

- `docs/workflows/release.yml`

To enable automated releases, copy it to:

- `.github/workflows/release.yml`

Then push a tag matching `v*` and GitHub Actions will build and publish the release.

> Note: creating/updating `.github/workflows/*` may require a token with the `workflow` scope.

## Option B: manual release

1. Update `CHANGELOG.md` (move items from `[Unreleased]` into the new version section).
2. Commit changes.
3. Create and push a tag:

```bash
git tag v0.1.1
git push origin v0.1.1
```

4. Build and publish release assets (example using `gh`):

```bash
VERSION=v0.1.1
COMMIT=$(git rev-parse --short HEAD)
mkdir -p dist
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o dist/sasswall_linux_amd64 ./cmd/sasswall
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o dist/sasswall_linux_arm64 ./cmd/sasswall
(cd dist && sha256sum sasswall_linux_* > sha256sums.txt)

gh release create "${VERSION}" dist/sasswall_linux_amd64 dist/sasswall_linux_arm64 dist/sha256sums.txt \
  --repo OWNER/REPO -t "${VERSION}" -n "Sasswall ${VERSION}"
```

## Version metadata

Release builds should embed:

- `main.version` = tag name (e.g., `v0.1.1`)
- `main.commit` = short commit SHA

Validate:

```bash
./sasswall -version
```
