# Command Cookbook

Use these patterns when you need the CLI to do real work.

## Environment Discovery

- Current API host:
  `go run ./cmd/chilly settings get api-base-url --output json`
- Full local config:
  `go run ./cmd/chilly settings show --output json`
- Config file path:
  `go run ./cmd/chilly settings path --output json`

Fresh configs default to `https://api.binge.institute`. Saved local config can override that.

## Schema And Describe

- List all known command and procedure metadata:
  `go run ./cmd/chilly schema --output json`
- Inspect one command:
  `go run ./cmd/chilly schema command search --output json`
- Inspect one nested command:
  `go run ./cmd/chilly schema command "settings get" --output json`
- Inspect one procedure:
  `go run ./cmd/chilly schema procedure chill.v4.UserService/Search --output json`
- Describe a command without executing it:
  `go run ./cmd/chilly search --describe --output json`

## Version And Update

- Show build metadata:
  `go run ./cmd/chilly version --output json`
- Check whether a newer release exists:
  `go run ./cmd/chilly self-update --check --output json`
- Install the latest release over the current binary:
  `go run ./cmd/chilly self-update`
- Install a specific release:
  `go run ./cmd/chilly self-update --version v0.1.0`

## Authentication

- Interactive browser-assisted login:
  `go run ./cmd/chilly auth login`
- Manual browser opening:
  `go run ./cmd/chilly auth login --no-browser`
- Existing token:
  `go run ./cmd/chilly auth login --token <token>`
- Logout:
  `go run ./cmd/chilly auth logout --output json`
- Verify current auth:
  `go run ./cmd/chilly whoami --output json`

## Read Commands

- Search:
  `go run ./cmd/chilly search --query "blade runner" --output json`
- Search with a specific indexer:
  `go run ./cmd/chilly search --query "blade runner" --indexer-id <id> --output json`
- User profile:
  `go run ./cmd/chilly whoami --output json`
- User indexers:
  `go run ./cmd/chilly user indexers --output json`
- Top movies:
  `go run ./cmd/chilly list-top-movies --output json`
- User settings:
  `go run ./cmd/chilly user settings get --output json`

## Mutating Commands

- Add transfer:
  `go run ./cmd/chilly add-transfer --url "magnet:?xt=..." --output json`
- Same operation through the nested command:
  `go run ./cmd/chilly user transfer add --url "magnet:?xt=..." --output json`
- Replace user settings with a full JSON payload:
  `go run ./cmd/chilly user settings set --json '{"showTopMovies":true}' --output json`

`user settings set` currently expects a full JSON object, not a partial patch.

## Output Discipline

- For automation, always request `--output json`.
- Human-facing notices may appear on `stderr`.
- Command data is intended to be parsed from `stdout`.
- In JSON mode, failures are emitted as a single JSON envelope on `stderr`.
