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

## Quickstart

### Prompt for Agents

```text
Use `chilly` to interact with chill.institute.

Repository:
https://github.com/chill-institute/cli

Before using the CLI:
1. Read the install instructions in the repo README:
   https://github.com/chill-institute/cli/blob/main/README.md
2. Download and follow the CLI usage skill:
   https://raw.githubusercontent.com/chill-institute/cli/main/skills/chilly-cli/SKILL.md

Install `chilly` if it is not already on PATH.

After install, run:
chilly doctor --output json

Then follow the skill for command usage, output conventions, and safe mutation patterns
```

### Commands for Humans

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
