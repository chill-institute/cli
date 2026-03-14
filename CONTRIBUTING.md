# Contributing

Thanks for contributing to `chill-institute/cli`.

## Setup

Install the toolchain:

```bash
mise install
mise run hooks
```

Build the CLI locally:

```bash
go build ./cmd/chilly
```

Run from source while developing:

```bash
go run ./cmd/chilly version --output json
```

## Validation

Run the full repo checks before opening or updating a pull request:

```bash
mise run verify
```

Format Go files when needed:

```bash
mise run fmt
```

Other shared development commands:

```bash
mise run fmt:check
mise run lint
mise run coverage
mise run security
```

The repository ships git hooks in `.githooks/`:

- `pre-commit` formats staged Go files with `goimports` and `gofmt`
- `pre-push` runs `mise run verify`

## Development Notes

- `chilly` is the command-line client for `chill.institute`.
- Prefer explicit, scriptable CLI behavior.
- Keep human output and machine-readable output both in mind.
- Keep `skills/chilly-cli/` in sync when install flows, auth behavior, schema output, or command surfaces change.

## Pull Requests

- Keep commands and flags stable unless a change is clearly justified.
- Update docs or examples when UX changes.
- Add or update tests when behavior changes.
