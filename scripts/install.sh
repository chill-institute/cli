#!/usr/bin/env bash

set -euo pipefail

repo="chill-institute/chill-institute-cli"
binary_name="chilly"
install_dir="${INSTALL_DIR:-/usr/local/bin}"
requested_version="${1:-${VERSION:-}}"
github_api_base="${GITHUB_API_BASE_URL:-https://api.github.com}"
github_download_base="${GITHUB_DOWNLOAD_BASE_URL:-https://github.com}"

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

checksum_cmd() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{ print $1 }'
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{ print $1 }'
    return
  fi

  printf 'missing checksum tool: need sha256sum or shasum\n' >&2
  exit 1
}

if [[ -z "$requested_version" ]]; then
  requested_version="$(
    curl -fsSL "${github_api_base}/repos/${repo}/releases/latest" | \
      awk -F '"' '/"tag_name":/ { print $4; exit }'
  )"
fi

if [[ -z "$requested_version" ]]; then
  printf 'failed to resolve release version\n' >&2
  exit 1
fi

release_tag="${requested_version}"
asset_version="${release_tag#v}"

archive_name="${binary_name}_${asset_version}_${os}_${arch}.tar.gz"
checksums_name="checksums.txt"
download_url="${github_download_base}/${repo}/releases/download/${release_tag}/${archive_name}"
checksums_url="${github_download_base}/${repo}/releases/download/${release_tag}/${checksums_name}"

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

printf '==> downloading %s\n' "$download_url"
curl -fsSL "$download_url" -o "${tmp_dir}/${archive_name}"

printf '==> downloading %s\n' "$checksums_url"
curl -fsSL "$checksums_url" -o "${tmp_dir}/${checksums_name}"

expected_checksum="$(
  awk -v name="${archive_name}" '
    {
      file=$NF
      sub(/^\*/, "", file)
      if (file == name) {
        print $1
        exit
      }
    }
  ' "${tmp_dir}/${checksums_name}"
)"
if [[ -z "${expected_checksum}" ]]; then
  printf 'failed to resolve checksum for %s\n' "${archive_name}" >&2
  exit 1
fi

actual_checksum="$(checksum_cmd "${tmp_dir}/${archive_name}")"
if [[ "${actual_checksum}" != "${expected_checksum}" ]]; then
  printf 'checksum mismatch for %s\n' "${archive_name}" >&2
  printf '  got:  %s\n' "${actual_checksum}" >&2
  printf '  want: %s\n' "${expected_checksum}" >&2
  exit 1
fi

printf '==> extracting %s\n' "$archive_name"
tar -xzf "${tmp_dir}/${archive_name}" -C "$tmp_dir"

mkdir -p "$install_dir"
install -m 0755 "${tmp_dir}/${binary_name}" "${install_dir}/${binary_name}"

printf 'installed %s %s to %s/%s\n' "$binary_name" "$release_tag" "$install_dir" "$binary_name"
