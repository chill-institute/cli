#!/usr/bin/env bash

set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"

cd "$repo_root"
git config core.hooksPath .githooks

printf 'git hooks configured: %s/.githooks\n' "$repo_root"
