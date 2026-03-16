# Security

If you believe you have found a security issue in the `chilly` CLI, please report it privately.

## Contact

- email: [chill-institute@proton.me](mailto:chill-institute@proton.me)

Private reports are preferred for security issues.

If you are unsure whether something is sensitive, feel free to email first.

## Scope

Useful reports include issues involving:

- auth token exposure or unsafe local credential handling
- updater, installer, or release integrity bypasses
- privilege escalation through config, profiles, or filesystem behavior
- command injection, path traversal, or unsafe external input handling
- private data exposure in CLI output, logs, or config files

## Guidelines

- test only against accounts and data you control
- avoid anything destructive, disruptive, or automated at scale
- do not target infrastructure outside the published CLI release, Homebrew tap, or documented API flows

## Supported Versions

Security fixes are applied on a best-effort basis to the latest released CLI version and the latest code on `main`.

`chilly` is currently in beta and defaults to the staging backend at `https://api.binge.institute`.

## Release Integrity

- published archives include `checksums.txt`
- `scripts/install.sh` verifies archive checksums before installation
- `chilly self-update` verifies archive checksums before replacing the current executable
- GitHub Actions publishes release artifact attestations for released archives

## Verify A Release

```bash
VERSION="$(gh release view --repo chill-institute/chill-institute-cli --json tagName -q .tagName)"
ARCHIVE="chilly_${VERSION#v}_darwin_arm64.tar.gz"

gh release download "$VERSION" --repo chill-institute/chill-institute-cli --pattern "$ARCHIVE"
gh attestation verify "$ARCHIVE" --repo chill-institute/chill-institute-cli
```

Adjust the archive name for your platform when needed.

## Disclosure

Please give us a reasonable chance to investigate and fix the issue before sharing details publicly.

There is no bug bounty program right now, but thoughtful reports are appreciated.
