# Read Reference

Use this reference for read-only workflows against the hosted API.

## Rules

- Prefer `--output json`
- Prefer `--fields` whenever the command supports it.
- Start wide enough to discover IDs, then rerun narrowly with selected fields.
- Parse only `stdout` Progress indicators and notices may appear on `stderr`
- Prefer top-level canonical commands over nested aliases when both exist.

## Search

- Discover indexers first: `chilly user indexers --fields indexers.id,indexers.name,indexers.tags --output json`
- Run one scoped search at a time: `chilly search --query "dune" --indexer-id yts --fields results.title --output json`
- Treat `--indexer-id` as an opaque ID. Inputs containing `/`, `?`, `#`, `%`, or traversal-like `..` are rejected locally.
- When reading `user indexers`, expect `indexers.status` to be a tri-state contract. `INDEXER_STATUS_DEGRADED` means the provider is partially working and should not be treated as fully down.

## Common Reads

- Authenticated profile: `chilly whoami --fields username,email --output json`
- Top movies: `chilly list-top-movies --fields movies.title --output json`
- Transfer details: `chilly get-transfer 42 --fields transfer.status,transfer.percentDone --output json`
- Hosted user settings: `chilly user settings get --fields showTopMovies,sortBy --output json`
- Current download folder: `chilly user download-folder --fields folder.id,folder.name --output json`
- Folder tree slice: `chilly user folder get 0 --fields parent.name,files.name --output json`

## Environment Reads

- Current host: `chilly settings get api-base-url --fields value --output json`
- Config path: `chilly settings path --fields path --output json`
- Local state overview: `chilly doctor --fields auth.status,config.profile,config.api_base_url --output json`
