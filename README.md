# chill-institute/cli

![chill.institute cli](https://binge.institute/banner.png)

Command line client for [chill.institute](https://chill.institute), your favorite [put.io](https://put.io) extension since 2018.

## Install

### Homebrew

```bash
brew install chill-institute/tap/chilly

chilly version --output json
```

### Direct Install

```bash
curl -fsSL https://raw.githubusercontent.com/chill-institute/cli/main/scripts/install.sh | bash

# optional pinned version
# curl -fsSL https://raw.githubusercontent.com/chill-institute/cli/main/scripts/install.sh | bash -s -- v0.1.0

# optional custom install directory
# curl -fsSL https://raw.githubusercontent.com/chill-institute/cli/main/scripts/install.sh | INSTALL_DIR="$HOME/.local/bin" bash

chilly version --output json
```

## Update

```bash
# check whether a newer release exists
chilly self-update --check --output json

# install the latest release over the current binary
chilly self-update

# install a specific release
chilly self-update --version v0.1.0
```

## Shell Completion

```bash
# zsh
chilly completion zsh > "${fpath[1]}/_chilly"

# bash
mkdir -p ~/.local/share/bash-completion/completions
chilly completion bash > ~/.local/share/bash-completion/completions/chilly

# fish
mkdir -p ~/.config/fish/completions
chilly completion fish > ~/.config/fish/completions/chilly.fish
```

## Quickstart

```bash
# 1) login with the browser-assisted flow
chilly auth login

# 2) confirm the token works
chilly whoami --output json

# 3) inspect local config, auth, and environment health
chilly doctor --output json

# 4) inspect the local CLI contract
chilly schema --output json

# 5) search
chilly search --query "dune" --output json

# 6) use the default pretty mode for a readable terminal summary
chilly search --query "dune"

# 7) narrow a read response to the fields you need
chilly search --query "dune" --fields results.title --output json

# 8) add a transfer
chilly add-transfer --url "magnet:?xt=urn:btih:..." --output json

# 9) preview a mutation without executing it
chilly add-transfer --url "magnet:?xt=urn:btih:..." --dry-run --output json
```

## Common Commands

```bash
# auth
chilly auth login --no-browser
chilly auth login --token <token>
chilly auth logout --output json
chilly auth logout --dry-run --output json

# discovery
chilly doctor --output json
chilly doctor --fields auth.status,config.profile --output json
chilly schema command search --output json
chilly search --describe --output json
chilly completion zsh > "${fpath[1]}/_chilly"

# read shaping
chilly whoami --fields username,email --output json
chilly search --query "dune" --fields results.title,results.magnetLink --output json
chilly list-top-movies --fields movies.title --output json
chilly user settings get --fields showTopMovies,sortBy --output json

# settings
chilly settings path --output json
chilly settings path --profile dev --output json
chilly settings get api-base-url --output json
chilly settings set api-base-url https://api.binge.institute
chilly settings set api-base-url https://api.chill.institute
chilly settings set api-base-url https://api.chill.institute --dry-run --output json
chilly settings show --output json

# user commands
chilly list-top-movies --output json
chilly user settings get --output json
chilly user settings set show-top-movies true --output json
chilly user settings set sort-by title --dry-run --output json
chilly user download-folder --output json
chilly user download-folder set 42 --dry-run --output json
chilly user download-folder clear --dry-run --output json
chilly user folder get 0 --output json
chilly add-transfer --url "magnet:?xt=..." --dry-run --output json
chilly settings set api-base-url https://api.chill.institute --dry-run --output json
chilly user settings set --json '{"showTopMovies":true}' --output json
chilly user settings set --json '{"showTopMovies":true}' --dry-run --output json
```

## Config And Auth

Default config path follows XDG:

- default profile: `$XDG_CONFIG_HOME/chilly/config.json`
- named profiles: `$XDG_CONFIG_HOME/chilly/profiles/<profile>/config.json`
- fallback base dir: `~/.config`

Profile behavior:

- released binaries default to the `default` profile
- dev builds default to the `dev` profile, so source runs do not reuse your production config by accident
- override with `--profile <name>` or `CHILLY_PROFILE=<name>`
- `--config <path>` still wins when you need a fully custom location

Storage properties:

- config file permissions: `0600`
- config directory permissions: `0700`
- writes are atomic via temp-file replace

Fresh configs default to `https://api.binge.institute`. Point the CLI at `https://api.chill.institute` when you want production.

Examples:

- `chilly settings path --output json`
- `chilly --profile dev settings show --output json`
- `CHILLY_PROFILE=staging chilly auth login`

`chilly auth login` starts a temporary loopback HTTP server on `127.0.0.1`, opens the API OAuth start URL, and waits for the browser callback to hand the issued auth token back to the CLI.

The browser is still required for put.io authentication. `--no-browser` only disables automatic browser launch so you can open the printed login URL yourself.

## Agent-First Contract

`chilly schema` exposes local metadata for public commands and linked backend procedures used by the CLI.

Mutating commands that support `--dry-run` return a local preview of the request or config change instead of touching auth state, local config, or the API.
Read commands that support `--fields` return only the selected paths from the JSON response.
`chilly doctor` reports build, profile, config path, API base URL, and auth health in one place.
In default pretty mode, `whoami`, `search`, `list-top-movies`, `user settings get`, and `user indexers` render concise terminal summaries for humans.
`chilly user settings set` supports both full JSON replacement and one-field patch updates for common settings.
Where top-level and nested `user ...` commands overlap, the top-level command is the canonical surface and the nested command is an alias under the user namespace.

Examples:

- `chilly doctor --output json`
- `chilly schema --output json`
- `chilly schema command search --output json`
- `chilly schema procedure chill.v4.UserService/Search --output json`
- `chilly search --describe --output json`

In `--output json` mode:

- command results go to `stdout`
- failures are emitted as a single JSON envelope on `stderr`
- exit classes are `0` success, `2` usage, `3` auth, `4` API, `5` internal

## Docs

- [Architecture](./docs/ARCHITECTURE.md)

## Contributing

Please read the [contributing guide](./CONTRIBUTING.md).

## License

Licensed under the [MIT License](./LICENSE).
