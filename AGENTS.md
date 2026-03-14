# cli Agent Notes

## Purpose

`cli` is the `chill.institute` command line client and should evolve as an agent-first SDK surface, not only a human shell tool.

## Maintainer Philosophy

- Prefer machine-readable contracts first. New command behavior should have a stable JSON story before adding pretty output.
- Keep `stdout` for command results and `stderr` for prompts, progress, warnings, and recovery hints.
- Favor explicit, inspectable surfaces over hidden behavior. Schema/describe support, clear auth requirements, and narrow flags beat magical prompts.
- Design mutating commands so `--dry-run`, idempotence, and future field filtering can be added cleanly.
- Keep shared logic reusable. Put transport, auth, normalization, and response handling in deep modules that can later back CLI, agents, and MCP surfaces.
- Treat agent input as untrusted. Validate and normalize flags, JSON payloads, URLs, and file paths at the boundary.

## Validation

Run these before finishing meaningful changes:

- `go test ./...`
- `mise run verify`
- For command surface changes: `go run ./cmd/chilly <command> --help`

## Skills

- Repo-local skills live in [skills/](./skills/).
- The repo-local `chilly-cli` skill is for agents using the CLI, not for maintainers editing the CLI.
- Keep build philosophy and contributor instructions in this file, not in the usage skill.
- As command surfaces, auth flows, defaults, or output contracts change, update the consumer-facing `chilly-cli` skill in the same pass so downstream agents stay aligned with the real SDK behavior.
