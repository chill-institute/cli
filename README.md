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

## Quickstart

```bash
# 1) login with the browser-assisted flow
chilly auth login

# 2) confirm the token works
chilly whoami --output json

# 3) inspect the local CLI contract
chilly schema --output json

# 4) search
chilly search --query "dune" --output json

# 5) add a transfer
chilly add-transfer --url "magnet:?xt=urn:btih:..." --output json
```

## Common Commands

```bash
# auth
chilly auth login --no-browser
chilly auth login --token <token>
chilly auth logout --output json

# discovery
chilly schema command search --output json
chilly search --describe --output json

# settings
chilly settings path --output json
chilly settings get api-base-url --output json
chilly settings set api-base-url https://api.binge.institute
chilly settings set api-base-url https://api.chill.institute
chilly settings show --output json

# user commands
chilly list-top-movies --output json
chilly user settings get --output json
chilly user settings set --json '{"showTopMovies":true}' --output json
```

## Config And Auth

Default config path follows XDG:

- `$XDG_CONFIG_HOME/chilly/config.json`
- fallback: `~/.config/chilly/config.json`

Storage properties:

- config file permissions: `0600`
- config directory permissions: `0700`
- writes are atomic via temp-file replace

Fresh configs default to `https://api.binge.institute`. Point the CLI at `https://api.chill.institute` when you want production.

`chilly auth login` starts a temporary loopback HTTP server on `127.0.0.1`, opens the API OAuth start URL, and waits for the browser callback to hand the issued auth token back to the CLI.

The browser is still required for put.io authentication. `--no-browser` only disables automatic browser launch so you can open the printed login URL yourself.

## Agent-First Contract

`chilly schema` exposes local metadata for public commands and linked backend procedures used by the CLI.

Examples:

- `chilly schema --output json`
- `chilly schema command search --output json`
- `chilly schema procedure chill.v4.UserService/Search --output json`
- `chilly search --describe --output json`

In `--output json` mode:

- command results go to `stdout`
- failures are emitted as a single JSON envelope on `stderr`
- exit classes are `0` success, `2` usage, `3` auth, `4` API, `5` internal

## Contributing

Please read the [contributing guide](./CONTRIBUTING.md).

## Docs

- [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)

## License

Licensed under the [MIT License](./LICENSE).
