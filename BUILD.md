# Build & Release Guide

## Prerequisites

- Go 1.16+
- `make`
- Bluetooth adapter (for running the binary)
- `zip` and `sha256sum` (for `make release`; on macOS use `shasum -a 256` or install coreutils)
- `gh` CLI authenticated to GitHub (for manual releases and workflow monitoring)

---

## Local development build

```bash
make build
# binary lands at build/aranet4-go
```

Check the version embedded at build time:

```bash
./build/aranet4-go -v
# aranet4-go v1.2.0 (commit: abc1234, built: 2026-09-09_12:00:00)
```

---

## Running tests

```bash
make test
```

---

## Creating a release

### 1. Make sure the repo is clean

```bash
git status           # nothing uncommitted
git log --oneline -5 # confirm HEAD is what you want to release
```

### 2. Bump the version in the Makefile

Edit `Makefile` and update `VERSION`:

```makefile
VERSION?=1.2.0
```

Follow semver: `MAJOR.MINOR.PATCH`

- Patch bump for bug fixes (`1.1.1` → `1.1.2`)
- Minor bump for new features (`1.1.x` → `1.2.0`)
- Major bump for breaking changes (`1.x.x` → `2.0.0`)

### 3. Commit the version bump, tag, and push

```bash
git add Makefile
git commit -m "Release v1.2.0"
git tag v1.2.0
git push origin main
git push origin v1.2.0
```

Pushing the tag triggers the GitHub Actions release workflow automatically.

### 4. Watch the workflow

```bash
gh run watch
```

The workflow will:
1. Run `go test ./...`
2. Build Linux (amd64 + arm64) on Ubuntu
3. Build macOS (amd64 + arm64) on macOS with CGO enabled (required for CoreBluetooth)
4. Package archives as `.tar.gz`
5. Generate `checksums.txt` with SHA256 hashes
6. Create a GitHub release with auto-generated notes and upload all artifacts

### 5. Verify the release

```bash
gh release view v1.2.0
```

---

## Manual release (without GitHub Actions)

If you need to cut a release locally, ensure the Makefile VERSION is already bumped and committed, then:

```bash
git tag v1.2.0   # if not already tagged

make release
# artifacts land in dist/

gh release create v1.2.0 \
  --title "Release v1.2.0" \
  --generate-notes \
  dist/*
```

---

## Deleting a bad release and re-triggering

If a release workflow fails or you need to redo it:

```bash
# Delete the GitHub release (keeps the tag)
gh release delete v1.2.0 --yes

# Delete the tag locally and on remote
git tag -d v1.2.0
git push origin --delete v1.2.0

# Fix whatever was wrong, then re-tag and push
git tag v1.2.0
git push origin v1.2.0
```

---

## GitHub CLI quick reference

### Releases

```bash
# List all releases
gh release list

# View a specific release (shows assets and notes)
gh release view v1.2.0

# Download a release asset
gh release download v1.2.0 --pattern "aranet4-go-*-linux-amd64.tar.gz"

# Edit release notes after the fact
gh release edit v1.2.0 --notes "Updated release notes"

# Delete a release (tag stays)
gh release delete v1.2.0 --yes
```

### Workflow runs

```bash
# List recent runs
gh run list

# List runs for the release workflow only
gh run list --workflow release.yml

# Watch a run in real time (run ID from gh run list)
gh run watch <run-id>

# View logs for a specific run
gh run view <run-id> --log

# View only the failed steps
gh run view <run-id> --log-failed

# Re-run a failed workflow
gh run rerun <run-id>

# Re-run only the failed jobs (skips passing jobs)
gh run rerun <run-id> --failed
```

### Tags

```bash
# List all tags
git tag -l

# Delete a remote tag (e.g. after a bad release)
git push origin --delete v1.2.0

# Push all local tags at once
git push origin --tags
```

### Repo inspection

```bash
# Open the repo in browser
gh repo view --web

# View open issues
gh issue list

# View open pull requests
gh pr list
```

---

## Notes

- The version is set manually in the Makefile (`VERSION?=1.1.1`) and must be
  bumped before tagging a release. The git tag and Makefile VERSION should match.
- The macOS build uses `CGO_ENABLED=1` because `go-ble/ble` uses Apple's
  CoreBluetooth (via the `cbgo` library), which requires CGO. Linux builds use
  `CGO_ENABLED=0` and are pure Go. Windows is not supported by `go-ble/ble`.
- The `DEFAULT-MAC-ADDR` file (gitignored) can bake a default MAC address into
  the binary at build time via ldflags.
- Linux binaries need `cap_net_raw,cap_net_admin` or `sudo` to access Bluetooth.
  The `make install` target sets these capabilities automatically.
