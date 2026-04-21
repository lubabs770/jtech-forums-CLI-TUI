#!/usr/bin/env sh

set -eu

REPO="${JTECHFORUMS_REPO:-lubabs770/jtech-forums-CLI-TUI}"
BIN_NAME="jtechforums"
INSTALL_DIR="${JTECHFORUMS_INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${JTECHFORUMS_VERSION:-latest}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

detect_os() {
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux) echo "linux" ;;
    darwin) echo "darwin" ;;
    mingw*|msys*|cygwin*) echo "windows" ;;
    *)
      echo "unsupported operating system: $os" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)
      echo "unsupported architecture: $arch" >&2
      exit 1
      ;;
  esac
}

release_url() {
  asset="$1"
  if [ "$VERSION" = "latest" ]; then
    echo "https://github.com/$REPO/releases/latest/download/$asset"
  else
    case "$VERSION" in
      v*) tag="$VERSION" ;;
      *) tag="v$VERSION" ;;
    esac
    echo "https://github.com/$REPO/releases/download/$tag/$asset"
  fi
}

main() {
  need_cmd curl
  need_cmd tar
  need_cmd mktemp

  os="$(detect_os)"
  arch="$(detect_arch)"
  asset="${BIN_NAME}-${os}-${arch}.tar.gz"
  url="$(release_url "$asset")"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT INT TERM

  archive="$tmpdir/$asset"

  echo "Downloading $url"
  curl -fsSL "$url" -o "$archive"

  mkdir -p "$INSTALL_DIR"
  tar -xzf "$archive" -C "$tmpdir"

  binary="$tmpdir/$BIN_NAME"
  if [ "$os" = "windows" ]; then
    binary="$tmpdir/${BIN_NAME}.exe"
  fi

  install_path="$INSTALL_DIR/$BIN_NAME"
  if [ "$os" = "windows" ]; then
    install_path="$INSTALL_DIR/${BIN_NAME}.exe"
  fi

  cp "$binary" "$install_path"
  chmod 755 "$install_path"

  echo "Installed $BIN_NAME to $install_path"
  case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
      echo "Add $INSTALL_DIR to your PATH to run $BIN_NAME directly." >&2
      ;;
  esac
}

main "$@"
