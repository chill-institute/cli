# chill-institute/cli

![chill.institute cli](https://binge.institute/banner.png)

CLI client for [chill.institute](https://chill.institute), your favorite [put.io](https://put.io) extension since 2018.

## Install

```bash
brew install chill-institute/tap/chilly
chilly version
```

```bash
curl -fsSL https://raw.githubusercontent.com/chill-institute/cli/main/scripts/install.sh | bash
chilly version
```

> [!NOTE]
> `chilly` is currently in beta and uses the staging backend at `api.binge.institute`

## Agent-First Contract

`chilly` is designed for both humans and agents.

- `chilly schema` and `--describe` expose the local command contract.
- `--fields` narrows read responses to the paths you need.
- `--dry-run` previews supported mutations without touching local config or the API.
- `chilly doctor` reports build, profile, config path, API base URL, and auth health in one place.
- In `--output json` mode, results go to `stdout` and failures go to `stderr` as one JSON envelope.
- In pretty mode, core read commands render concise summaries for humans.

```bash
chilly schema --output json
chilly search --describe --output json
chilly doctor --fields auth.status,config.profile --output json
chilly search --query "dune" --fields results.title --output json
chilly add-transfer --url "magnet:?xt=urn:btih:..." --dry-run --output json
```

## Quickstart

```bash
chilly auth login
chilly doctor --output json
chilly whoami --output json
chilly search --query "dune"
chilly add-transfer --url "magnet:?xt=urn:btih:..." --dry-run --output json
```

Released binaries use the `default` profile; dev builds default to `dev` so source runs do not reuse production config by accident.

## Verify Release

```bash
VERSION="$(gh release view --repo chill-institute/cli --json tagName -q .tagName)"
ARCHIVE="chilly_${VERSION#v}_darwin_arm64.tar.gz"

gh release download "$VERSION" --repo chill-institute/cli --pattern "$ARCHIVE"
gh attestation verify "$ARCHIVE" --repo chill-institute/cli
```

The install script verifies release checksums. Use GitHub attestation verification when you also want provenance from the release workflow.

## Docs

- [Architecture](./docs/ARCHITECTURE.md)

## Contributing

Please read the [contributing guide](./CONTRIBUTING.md).

## License

Licensed under the [MIT License](./LICENSE).
