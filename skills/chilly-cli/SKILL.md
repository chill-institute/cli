---
name: chilly-cli
description: Use `chilly` to interact with chill.institute from the terminal. Trigger this skill when an agent needs to authenticate, inspect the current user, search, list top movies, add a transfer, read or update user settings, or switch between hosted API environments with the local CLI. Covers the safest command patterns, JSON output usage, and the current auth flow.
---

# Chilly CLI

Use `chilly` as the local command-line entrypoint for chill.institute. Prefer `--output json` for any agent workflow that will parse or transform results.

## Quick Start

1. Assume `chilly` is installed and available on `PATH`.
2. Use direct binary commands such as `chilly whoami --output json`.
3. Fall back to `go run ./cmd/chilly ...` only when working from source in a maintainer checkout.
4. Add `--output json` whenever the result will feed another tool or decision.
5. Omit `--output json` when a human wants the built-in terminal summary from `whoami`, `search`, `list-top-movies`, `user settings get`, or `user indexers`.
6. Use `settings get api-base-url` before assuming which hosted environment is active.
7. Use `--profile <name>` when you need an isolated config path instead of reusing the default profile.
8. Use `doctor` before debugging auth, profile, config path, or environment mismatches.
9. Use `schema` or `--describe` when you need to inspect the local CLI contract before running a command.
10. Use `version` and `self-update --check` when you need release provenance before proposing an upgrade.
11. Use `--dry-run` on supported mutating commands when you need to preview a request safely.
12. Use `--fields` on supported read commands when you only need a stable subset of the JSON response.
13. Use `completion` when you need shell integration on a human workstation.

## Auth

- Default interactive login:
  `chilly auth login`
- Print the login URL without auto-opening a browser:
  `chilly auth login --no-browser`
- Store an existing token directly:
  `chilly auth login --token <token>`
- Verify the current login:
  `chilly whoami --output json`

The current fresh-config default is `https://api.binge.institute`. Existing local configs may already point somewhere else.

## Discovery

- List local command and procedure metadata:
  `chilly schema --output json`
- Show one command schema:
  `chilly schema command search --output json`
- Show one procedure schema:
  `chilly schema procedure chill.v4.UserService/Search --output json`
- Describe a command without executing it:
  `chilly search --describe --output json`
- Show installed build metadata:
  `chilly version --output json`
- Check whether a newer release exists:
  `chilly self-update --check --output json`
- Generate zsh completions:
  `chilly completion zsh`

## Common Commands

- Show current API host:
  `chilly settings get api-base-url --output json`
- Show current config path and profile:
  `chilly settings path --output json`
- Use an isolated dev profile:
  `chilly --profile dev settings show --output json`
- Run full local diagnostics:
  `chilly doctor --output json`
- Point to staging:
  `chilly settings set api-base-url https://api.binge.institute`
- Point to production:
  `chilly settings set api-base-url https://api.chill.institute`
- Preview an API host change without saving it:
  `chilly settings set api-base-url https://api.chill.institute --dry-run --output json`
- Search:
  `chilly search --query "dune" --output json`
- Search with a smaller response:
  `chilly search --query "dune" --fields results.title --output json`
- List top movies:
  `chilly list-top-movies --output json`
- List only movie titles:
  `chilly list-top-movies --fields movies.title --output json`
- Add transfer:
  `chilly add-transfer --url "magnet:?xt=..." --output json`
- Preview transfer request without executing it:
  `chilly add-transfer --url "magnet:?xt=..." --dry-run --output json`
- Preview logout without clearing the saved token:
  `chilly auth logout --dry-run --output json`
- Read user settings:
  `chilly user settings get --output json`
- Read only selected settings:
  `chilly user settings get --fields showTopMovies,sortBy --output json`
- Patch one setting:
  `chilly user settings set show-top-movies true --output json`
- Preview one patched setting:
  `chilly user settings set sort-by title --dry-run --output json`
- Show the active download folder:
  `chilly user download-folder --output json`
- Preview a download folder change:
  `chilly user download-folder set 42 --dry-run --output json`
- Preview clearing the download folder:
  `chilly user download-folder clear --dry-run --output json`
- Inspect one folder by id:
  `chilly user folder get 0 --output json`
- Preview full user settings update:
  `chilly user settings set --json '{"showTopMovies":true}' --dry-run --output json`
- Show build metadata:
  `chilly version --output json`

Read `references/commands.md` for a fuller command cookbook and current gotchas.

## Safety Rules

- Prefer `--output json` for automation.
- Prefer `--fields` when a command supports it and you only need a subset of the payload.
- Check the active API base URL before mutating anything.
- Use `doctor` when auth, config path, or profile state looks inconsistent.
- Prefer `--dry-run` before a mutation when you only need the request shape or want a safe preview.
- Prefer the single-field `user settings set <field> <value>` form for routine settings changes, and keep `--json` for full replacement.
- Expect prompts and browser-login hints on `stderr`; parse only `stdout`.
- Expect failures in `--output json` mode to appear as a single JSON envelope on `stderr`.
- Use `whoami` after auth changes when you need a positive confirmation that the token works.
- Prefer top-level commands like `search`, `whoami`, `list-top-movies`, and `add-transfer`; use nested `user ...` aliases only when namespacing helps.
