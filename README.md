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

## Quickstart

### Prompt for Agents

```text
Use `chilly` to interact with chill.institute from the terminal

Repository:
https://github.com/chill-institute/chill-institute-cli

Read and follow this usage skill before operating the CLI:
https://raw.githubusercontent.com/chill-institute/chill-institute-cli/main/skills/chilly-cli/SKILL.md

When only one workflow is relevant, follow the progressive-disclosure references linked from that root skill instead of loading unrelated guidance.

If `chilly` is not already on PATH, install it by following the repo README:
https://github.com/chill-institute/chill-institute-cli/blob/main/README.md

After install, run:
chilly doctor --output json

If auth is missing, start the hosted web token flow:
chilly auth login

The command prints this page, asks the user to copy the setup token, and waits for it to be pasted back:
https://chill.institute/auth/cli-token

If you already have the token and want a non-interactive path, use:
chilly auth login --token <token>

After setup, continue with the requested task instead of stopping after install or doctor output
```

Treat the agent as an untrusted operator: prefer `--output json`, parse only `stdout`, use `--fields` to narrow reads, and use `--dry-run` before mutations.

### Commands for Humans

```bash
chilly auth login
chilly doctor --output json
chilly whoami --output json
chilly search --query "dune"
chilly version --fields version --output json
chilly add-transfer --url "magnet:?xt=urn:btih:..." --dry-run --output json
printf '{"url":"magnet:?xt=urn:btih:..."}' | chilly add-transfer --json @- --dry-run --output json
printf '{"token":"token-from-setup","skip_verify":true}' | chilly auth login --json @- --dry-run --output json
printf '{"key":"api-base-url","value":"https://api.chill.institute"}' | chilly settings set --json @- --dry-run --output json
chilly schema command search --fields id,linked_procedure --output json
chilly self-update --json '{"check":true}' --output json
```

Released binaries use the `default` profile; dev builds default to `dev` so source runs do not reuse production config by accident.

When `stdout` is not a TTY, `chilly` now defaults to compact JSON for command results unless `--output` is set explicitly. That keeps piped and agent-driven runs machine-readable without changing the interactive terminal summaries humans see by default.

## Docs

- [Architecture](./docs/ARCHITECTURE.md)
- [Security](./SECURITY.md)
- [Contributing](./CONTRIBUTING.md)

## Contributing

Please read the [contributing guide](./CONTRIBUTING.md).

## License

Licensed under the [MIT License](./LICENSE).
