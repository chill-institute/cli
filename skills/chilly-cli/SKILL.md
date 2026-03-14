---
name: chilly-cli
description: Use `chilly` to interact with chill.institute from the terminal. Trigger this skill when an agent needs to authenticate, inspect the current user, search, list top movies, add a transfer, read or update user settings, or switch between hosted API environments with the local CLI. Covers the safest command patterns, JSON output usage, and the current auth flow.
---

# Chilly CLI

Use `chilly` as the local command-line entrypoint for chill.institute. Prefer `--output json` for any agent workflow that will parse or transform results.

## Quick Start

1. Run commands from the `cli` repo root.
2. Prefer `go run ./cmd/chilly ...` unless the binary is already built and trusted.
3. Add `--output json` whenever the result will feed another tool or decision.
4. Use `settings get api-base-url` before assuming which hosted environment is active.
5. Use `schema` or `--describe` when you need to inspect the local CLI contract before running a command.
6. Use `version` and `self-update --check` when you need release provenance before proposing an upgrade.

## Auth

- Default interactive login:
  `go run ./cmd/chilly auth login`
- Print the login URL without auto-opening a browser:
  `go run ./cmd/chilly auth login --no-browser`
- Store an existing token directly:
  `go run ./cmd/chilly auth login --token <token>`
- Verify the current login:
  `go run ./cmd/chilly whoami --output json`

The current fresh-config default is `https://api.binge.institute`. Existing local configs may already point somewhere else.

## Discovery

- List local command and procedure metadata:
  `go run ./cmd/chilly schema --output json`
- Show one command schema:
  `go run ./cmd/chilly schema command search --output json`
- Show one procedure schema:
  `go run ./cmd/chilly schema procedure chill.v4.UserService/Search --output json`
- Describe a command without executing it:
  `go run ./cmd/chilly search --describe --output json`
- Show installed build metadata:
  `go run ./cmd/chilly version --output json`
- Check whether a newer release exists:
  `go run ./cmd/chilly self-update --check --output json`

## Common Commands

- Show current API host:
  `go run ./cmd/chilly settings get api-base-url --output json`
- Point to staging:
  `go run ./cmd/chilly settings set api-base-url https://api.binge.institute`
- Point to production:
  `go run ./cmd/chilly settings set api-base-url https://api.chill.institute`
- Search:
  `go run ./cmd/chilly search --query "dune" --output json`
- List top movies:
  `go run ./cmd/chilly list-top-movies --output json`
- Add transfer:
  `go run ./cmd/chilly add-transfer --url "magnet:?xt=..." --output json`
- Read user settings:
  `go run ./cmd/chilly user settings get --output json`
- Show build metadata:
  `go run ./cmd/chilly version --output json`

Read `references/commands.md` for a fuller command cookbook and current gotchas.

## Safety Rules

- Prefer `--output json` for automation.
- Check the active API base URL before mutating anything.
- Expect prompts and browser-login hints on `stderr`; parse only `stdout`.
- Expect failures in `--output json` mode to appear as a single JSON envelope on `stderr`.
- Use `whoami` after auth changes when you need a positive confirmation that the token works.
