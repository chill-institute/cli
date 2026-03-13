# `chill-institute/cli`

`chilly` is the Go CLI for `chill.institute`.

- Functional command mode: Cobra commands with `--output json` for automation or `--output pretty` for humans.

## Build

```bash
go build ./cmd/chilly
```

## Verification

```bash
mise install
mise run verify
```

## Quick start

```bash
# interactive login flow (opens web auth, then prompts for token)
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
- fallback: `~/.config/chilly/config.json` (platform equivalent via `os.UserConfigDir`)

File contents:

```json
{
  "api_base_url": "http://localhost:8080",
  "auth_token": "..."
}
```
