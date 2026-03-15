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

gofmt_files=()
while IFS= read -r file; do
  if [[ -n "$file" ]]; then
    gofmt_files+=("$file")
  fi
done < <(gofmt -l "${go_files[@]}")

goimports_files=()
while IFS= read -r file; do
  if [[ -n "$file" ]]; then
    goimports_files+=("$file")
  fi
done < <(go tool goimports -l "${go_files[@]}")

if [[ "${#gofmt_files[@]}" -eq 0 && "${#goimports_files[@]}" -eq 0 ]]; then
  printf '==> Go formatting is clean\n'
  exit 0
fi

if [[ "${#gofmt_files[@]}" -gt 0 ]]; then
  printf 'gofmt needs to be run on:\n' >&2
  printf '  %s\n' "${gofmt_files[@]}" >&2
fi

if [[ "${#goimports_files[@]}" -gt 0 ]]; then
  printf 'goimports needs to be run on:\n' >&2
  printf '  %s\n' "${goimports_files[@]}" >&2
fi

exit 1
