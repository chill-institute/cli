# Contributing

Thanks for contributing to `chill-institute/cli`.

## Setup

Install the toolchain:

```bash
mise install
```

Build the CLI locally:

```bash
go build ./cmd/chilly
```

## Validation

Run the full repo checks before opening or updating a pull request:

```bash
mise run verify
```

## Development Notes

- `chilly` is the command-line client for `chill.institute`.
- Prefer explicit, scriptable CLI behavior.
- Keep human output and machine-readable output both in mind.

## Pull Requests

- Keep commands and flags stable unless a change is clearly justified.
- Update docs or examples when UX changes.
- Add or update tests when behavior changes.
