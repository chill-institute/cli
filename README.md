# chill-institute-cli

![chill.institute cli](https://chill.institute/banner.png)

CLI client for [chill.institute](https://chill.institute), your favorite [put.io](https://put.io) extension since 2018.

## Install

```bash
brew install chill-institute/tap/chilly
chilly version
```

```bash
curl -fsSL https://raw.githubusercontent.com/chill-institute/chill-institute-cli/main/scripts/install.sh | bash
chilly version
```

> [!NOTE]
> `chilly` is currently in beta and uses the production backend at `api.chill.institute`

## Quickstart

### Prompt for Agents

```text
Use `chilly` to interact with chill.institute from the terminal

Repository:
https://github.com/chill-institute/chill-institute-cli

Read and follow this usage skill before operating the CLI:
https://raw.githubusercontent.com/chill-institute/chill-institute-cli/main/skills/chilly-cli/SKILL.md

If `chilly` is not already on PATH, install it by following the repo README:
https://github.com/chill-institute/chill-institute-cli/blob/main/README.md

After install, run:
chilly doctor --output json

If auth is missing and browser login is possible, run:
chilly auth login

If browser login is not possible on the CLI machine, ask the user to open this page in a signed-in browser and copy the token:
https://chill.institute/auth/cli-token

Then use:
chilly auth login --token <token>

After setup, continue with the requested task instead of stopping after install or doctor output
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

## Docs

- [Architecture](./docs/ARCHITECTURE.md)
- [Security](./SECURITY.md)

## Contributing

Please read the [contributing guide](./CONTRIBUTING.md).

## License

Licensed under the [MIT License](./LICENSE).
