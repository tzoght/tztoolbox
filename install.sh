#!/bin/sh
# tztoolbox bootstrap installer.
#
# This script downloads the latest `tzcli` release binary and runs
# `tzcli install` against the user's home directory. If a release binary
# is unavailable for the host OS/arch, it falls back to building from
# source with `go build` (requires Go to be installed).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/tzoght/tztoolbox/main/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/tzoght/tztoolbox/main/install.sh | sh -s -- --only cursor
#   curl -fsSL https://raw.githubusercontent.com/tzoght/tztoolbox/main/install.sh | sh -s -- --no-prune
#
# All flags after `--` are forwarded verbatim to `tzcli install`.

set -eu

REPO="${TZTOOLBOX_REPO:-tzoght/tztoolbox}"
INSTALL_DIR="${TZTOOLBOX_BIN_DIR:-${HOME}/.local/bin}"
SOURCE_DIR_HINT="${TZTOOLBOX_SOURCE_DIR:-}"

err() {
  printf '%s\n' "$*" >&2
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    err "Error: required command not found: $1"
    exit 1
  fi
}

detect_platform() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)
  case "$os" in
    darwin | linux) ;;
    *)
      err "Unsupported OS: $os. Falling back to source build."
      printf 'unsupported'
      return
      ;;
  esac
  case "$arch" in
    x86_64 | amd64) arch=amd64 ;;
    aarch64 | arm64) arch=arm64 ;;
    *)
      err "Unsupported arch: $arch. Falling back to source build."
      printf 'unsupported'
      return
      ;;
  esac
  printf '%s_%s' "$os" "$arch"
}

download_release_binary() {
  platform=$1
  asset="tzcli_${platform}.tar.gz"
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
  tmp=$(mktemp -d)
  printf '==> Downloading %s\n' "$url"
  if ! curl -fL --silent --show-error --output "${tmp}/${asset}" "$url"; then
    rm -rf "$tmp"
    return 1
  fi
  tar -xzf "${tmp}/${asset}" -C "$tmp"
  mkdir -p "$INSTALL_DIR"
  mv "${tmp}/tzcli" "${INSTALL_DIR}/tzcli"
  chmod +x "${INSTALL_DIR}/tzcli"
  rm -rf "$tmp"
  printf '==> Installed tzcli to %s\n' "${INSTALL_DIR}/tzcli"
  return 0
}

build_from_source() {
  need_cmd go
  need_cmd git
  if [ -n "$SOURCE_DIR_HINT" ] && [ -d "$SOURCE_DIR_HINT" ]; then
    src="$SOURCE_DIR_HINT"
  else
    src=$(mktemp -d)
    printf '==> Cloning %s\n' "$REPO"
    git clone --depth 1 "https://github.com/${REPO}.git" "$src" >/dev/null 2>&1
  fi
  printf '==> Building tzcli from source in %s\n' "$src"
  mkdir -p "$INSTALL_DIR"
  ( cd "$src" && go build -o "${INSTALL_DIR}/tzcli" ./cmd/tzcli )
  printf '==> Installed tzcli to %s\n' "${INSTALL_DIR}/tzcli"
}

ensure_path() {
  case ":${PATH:-}:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      err "Note: ${INSTALL_DIR} is not on your PATH. Add it to your shell rc:"
      err "  export PATH=\"${INSTALL_DIR}:\$PATH\""
      ;;
  esac
}

main() {
  need_cmd uname
  need_cmd curl
  need_cmd tar

  platform=$(detect_platform)
  if [ "$platform" = unsupported ] || ! download_release_binary "$platform"; then
    build_from_source
  fi

  ensure_path
  printf '==> Running: tzcli install %s\n' "$*"
  "${INSTALL_DIR}/tzcli" install "$@"
}

main "$@"
