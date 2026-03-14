# Command Cookbook

Use these patterns when you need the CLI to do real work.

## Environment Discovery

- Current API host:
  `chilly settings get api-base-url --output json`
- Full local config:
  `chilly settings show --output json`
- Config file path:
  `chilly settings path --output json`

Fresh configs default to `https://api.binge.institute`. Saved local config can override that.

## Schema And Describe

- List all known command and procedure metadata:
  `chilly schema --output json`
- Inspect one command:
  `chilly schema command search --output json`
- Inspect one nested command:
  `chilly schema command "settings get" --output json`
- Inspect one procedure:
  `chilly schema procedure chill.v4.UserService/Search --output json`
- Describe a command without executing it:
  `chilly search --describe --output json`

## Version And Update

- Show build metadata:
  `chilly version --output json`
- Check whether a newer release exists:
  `chilly self-update --check --output json`
- Install the latest release over the current binary:
  `chilly self-update`
- Install a specific release:
  `chilly self-update --version v0.1.0`

## Authentication

- Interactive browser-assisted login:
  `chilly auth login`
- Manual browser opening:
  `chilly auth login --no-browser`
- Existing token:
  `chilly auth login --token <token>`
- Logout:
  `chilly auth logout --output json`
- Verify current auth:
  `chilly whoami --output json`

## Read Commands

- Search:
  `chilly search --query "blade runner" --output json`
- Search with field selection:
  `chilly search --query "blade runner" --fields results.title,results.magnetLink --output json`
- Search with a specific indexer:
  `chilly search --query "blade runner" --indexer-id <id> --output json`
- User profile:
  `chilly whoami --output json`
- User profile with selected fields:
  `chilly whoami --fields username,email --output json`
- User indexers:
  `chilly user indexers --output json`
- Top movies:
  `chilly list-top-movies --output json`
- Top movies with selected fields:
  `chilly list-top-movies --fields movies.title --output json`
- User settings:
  `chilly user settings get --output json`
- User settings with selected fields:
  `chilly user settings get --fields showTopMovies,sortBy --output json`

## Mutating Commands

- Add transfer:
  `chilly add-transfer --url "magnet:?xt=..." --output json`
- Preview add-transfer without executing it:
  `chilly add-transfer --url "magnet:?xt=..." --dry-run --output json`
- Same operation through the nested command:
  `chilly user transfer add --url "magnet:?xt=..." --output json`
- Preview the nested transfer command:
  `chilly user transfer add --url "magnet:?xt=..." --dry-run --output json`
- Replace user settings with a full JSON payload:
  `chilly user settings set --json '{"showTopMovies":true}' --output json`
- Preview the full settings payload:
  `chilly user settings set --json '{"showTopMovies":true}' --dry-run --output json`

`user settings set` currently expects a full JSON object, not a partial patch.

## Output Discipline

- For automation, always request `--output json`.
- Human-facing notices may appear on `stderr`.
- Command data is intended to be parsed from `stdout`.
- In JSON mode, failures are emitted as a single JSON envelope on `stderr`.
