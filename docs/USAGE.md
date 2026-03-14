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
# interactive login flow
chilly auth login

# non-interactive login
chilly auth login --token <token>

# verify current auth
chilly whoami

# local CLI settings
chilly settings path
chilly settings get api-base-url
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
  "api_base_url": "http://localhost:8080",
  "auth_token": "..."
}
```
