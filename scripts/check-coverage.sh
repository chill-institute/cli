#!/usr/bin/env bash

set -euo pipefail

threshold="${1:-70}"
repo_root="$(git rev-parse --show-toplevel)"

cd "$repo_root"

printf '==> running Go tests with coverage (threshold: %s%%)\n' "$threshold"
go test ./... -covermode=atomic -coverprofile=coverage.out

total_coverage="$(
  go tool cover -func=coverage.out | awk '/^total:/ { gsub(/%/, "", $3); print $3 }'
)"

if [[ -z "$total_coverage" ]]; then
  printf 'failed to determine total coverage from coverage.out\n' >&2
  exit 1
fi

printf '==> total coverage: %s%%\n' "$total_coverage"

awk -v total="$total_coverage" -v threshold="$threshold" '
  BEGIN {
    if ((total + 0) < (threshold + 0)) {
      printf("coverage %.1f%% is below threshold %.1f%%\n", total + 0, threshold + 0) > "/dev/stderr"
      exit 1
    }
  }
'
