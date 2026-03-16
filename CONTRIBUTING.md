# Contributing

Thanks for contributing to `chill-institute-cli`.

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

## Release Flow

Normal CLI change flow:

1. Make the change.
2. Run `mise run verify`.
3. Merge or push to `main`.
4. GitHub Actions runs `semantic-release` on `main` and creates the next `vX.Y.Z` tag from conventional commits.
5. The same `main` release workflow runs GoReleaser, publishes release archives, creates the GitHub release, and updates the Homebrew tap.
6. The existing tag-based GoReleaser workflow remains available as a manual fallback for intentionally pushed tags.

Versioning notes:

- Releases are driven by conventional commits.
- `feat:` produces a minor release.
- `fix:`, `perf:`, `refactor:`, `docs:`, `test:`, `build:`, `ci:`, and `chore:` produce a patch release.
- Breaking changes produce a major release.
