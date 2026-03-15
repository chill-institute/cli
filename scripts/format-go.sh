#!/usr/bin/env bash

set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

go_files=()
while IFS= read -r file; do
  if [[ -n "$file" ]]; then
    go_files+=("$file")
  fi
done < <(rg --files -g '*.go')

if [[ "${#go_files[@]}" -eq 0 ]]; then
  exit 0
fi

printf '==> formatting Go files with goimports and gofmt\n'
go tool goimports -w "${go_files[@]}"
gofmt -w "${go_files[@]}"
