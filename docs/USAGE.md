# Usage

Common `chilly` workflows.

## Build And Verify

```bash
go build ./cmd/chilly
mise install
mise run verify
```

## Quick Start

```bash
# browser-assisted login flow
chilly auth login

# print the login URL without auto-opening a browser
chilly auth login --no-browser

# non-interactive login with an existing token
chilly auth login --token <token>

# verify current auth
chilly whoami

# inspect agent-facing metadata
chilly schema --output json
chilly schema command search --output json
chilly search --describe --output json

# inspect build metadata and release availability
chilly version --output json
chilly self-update --check --output json

# local CLI settings
chilly settings path
chilly settings get api-base-url
chilly settings set api-base-url https://api.binge.institute
chilly settings set api-base-url https://api.chill.institute
chilly settings show

# user commands
chilly search --query "dune"
chilly list-top-movies
chilly add-transfer --url "magnet:?xt=urn:btih:..."
chilly user settings get
chilly user settings set --json '{"showTopMovies":true}'

# logout
chilly auth logout
```

## Install And Update

Preferred install paths:

- Homebrew: `brew install chill-institute/tap/chilly`
- release installer: `curl -fsSL https://raw.githubusercontent.com/chill-institute/cli/main/scripts/install.sh | sh`

Update commands:

- `chilly self-update --check --output json`
- `chilly self-update`
- `chilly self-update --version v0.1.0`

## Local Settings

Default config path follows XDG:

- `$XDG_CONFIG_HOME/chilly/config.json`
- fallback: `~/.config/chilly/config.json`

Storage properties:

- config file permissions: `0600`
- config directory permissions: `0700`
- writes are atomic via temp-file replace

Config shape:

```json
{
  "api_base_url": "https://api.binge.institute",
  "auth_token": "..."
}
```

## Auth Flow

`chilly auth login` starts a temporary loopback HTTP server on `127.0.0.1`, opens the API OAuth start URL, and waits for the browser callback to hand the issued auth token back to the CLI.

The browser is still required for put.io authentication. `--no-browser` only disables automatic browser launch so you can open the printed login URL yourself.

Fresh configs default to `https://api.binge.institute`. Point the CLI at `https://api.chill.institute` when you want the production environment instead.

## Introspection

`chilly schema` exposes local metadata for public commands and backend procedures used by the CLI.

Examples:

- `chilly schema --output json`
- `chilly schema command search --output json`
- `chilly schema procedure chill.v4.UserService/Search --output json`
- `chilly search --describe --output json`

Quote command names with spaces when needed, for example:

- `chilly schema command "settings get" --output json`

## Exit Behavior

When `--output json` is active, command results still go to `stdout` and failures are emitted as a single JSON envelope on `stderr`.

Current exit classes:

- `0` success
- `2` usage or local validation error
- `3` auth missing or auth invalid
- `4` API error from the backend
- `5` internal runtime failure

## Quality Guardrails

Shared local and CI checks run through `mise`:

- `mise run fmt:check`
- `mise run lint`
- `mise run coverage`
- `mise run security`
- `mise run verify`
