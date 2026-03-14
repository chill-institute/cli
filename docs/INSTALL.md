# Install

Install `chilly` from a release artifact whenever possible.

## Homebrew

```bash
brew install chill-institute/tap/chilly
```

## Shell Installer

```bash
curl -fsSL https://raw.githubusercontent.com/chill-institute/cli/main/scripts/install.sh | sh
```

Optional overrides:

```bash
VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" sh ./scripts/install.sh
```

## From Source

```bash
mise install
go build ./cmd/chilly
./chilly version --output json
```

## Update

Check for a new release:

```bash
chilly self-update --check --output json
```

Install the newest release over the current binary:

```bash
chilly self-update
```

Install a specific release:

```bash
chilly self-update --version v0.1.0
```

## Release Maintainers

Tagged pushes matching `v*` run the release workflow.

Requirements:

- GitHub Actions secret `TAP_GITHUB_TOKEN` with push access to `chill-institute/homebrew-tap`
- clean `mise run verify`
- a tag such as `v0.1.0`
