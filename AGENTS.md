# CLI

`cli` is the `chill.institute` command-line client. Treat it as an agent-first SDK surface with a strong human CLI on top.

## Priorities

- Prefer machine-readable contracts first. New behavior should have a stable JSON story before it gets nicer human formatting.
- Keep `stdout` for command results and `stderr` for prompts, progress, warnings, and recovery hints.
- Favor explicit, inspectable surfaces. Clear auth requirements, schema/describe support, and narrow flags beat magical prompts.
- Design mutating commands so `--dry-run`, idempotence, and future field filtering can be added without reshaping the command.
- Keep transport, auth, normalization, and response handling reusable so the same core can back CLI, agents, and future MCP surfaces.
- Treat agent input as untrusted. Validate and normalize flags, JSON payloads, URLs, and file paths at the boundary.

## Contract Changes

- If a command surface, auth flow, default, or output contract changes, update the user-facing `chilly-cli` skill in [skills/](./skills/) in the same pass.
- Keep maintainer guidance in this file. Keep consumer usage guidance in the skill.

## Validation

- `go test ./...`
- `mise run verify`
- For command surface changes, also run `go run ./cmd/chilly <command> --help`.

## Release Flow

- Releases are cut from `main`, not by manually pushing tags.
- Keep release behavior aligned with [`.github/workflows/tag-release.yml`](./.github/workflows/tag-release.yml) and GoReleaser.
- If install docs or release assumptions change, update [README.md](./README.md) and [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) in the same pass.
