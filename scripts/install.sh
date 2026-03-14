#!/usr/bin/env bash

set -euo pipefail

repo="chill-institute/cli"
binary_name="chilly"
install_dir="${INSTALL_DIR:-/usr/local/bin}"
requested_version="${VERSION:-}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  darwin|linux)
    ;;
  *)
    printf 'unsupported operating system: %s\n' "$os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64)
    arch="amd64"
    ;;
  arm64|aarch64)
    arch="arm64"
    ;;
  *)
    printf 'unsupported architecture: %s\n' "$arch" >&2
    exit 1
    ;;
esac

if [[ -z "$requested_version" ]]; then
  requested_version="$(
    curl -fsSL "https://api.github.com/repos/${repo}/releases/latest" | \
      awk -F '"' '/"tag_name":/ { print $4; exit }'
  )"
fi

if [[ -z "$requested_version" ]]; then
  printf 'failed to resolve release version\n' >&2
  exit 1
fi

archive_name="${binary_name}_${requested_version}_${os}_${arch}.tar.gz"
download_url="https://github.com/${repo}/releases/download/${requested_version}/${archive_name}"

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

printf '==> downloading %s\n' "$download_url"
curl -fsSL "$download_url" -o "${tmp_dir}/${archive_name}"

printf '==> extracting %s\n' "$archive_name"
tar -xzf "${tmp_dir}/${archive_name}" -C "$tmp_dir"

mkdir -p "$install_dir"
install -m 0755 "${tmp_dir}/${binary_name}" "${install_dir}/${binary_name}"

printf 'installed %s %s to %s/%s\n' "$binary_name" "$requested_version" "$install_dir" "$binary_name"
