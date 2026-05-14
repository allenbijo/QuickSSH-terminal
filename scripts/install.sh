#!/usr/bin/env sh
# QuickSSH Terminal installer (macOS / Linux)
# Usage:  curl -fsSL https://github.com/allenbijo/QuickSSH-terminal/releases/latest/download/install.sh | sh
set -eu

REPO="allenbijo/QuickSSH-terminal"
BIN="quickssh-tui"

red()    { printf '\033[31m%s\033[0m\n' "$*"; }
green()  { printf '\033[32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }

uname_s=$(uname -s)
uname_m=$(uname -m)

case "$uname_s" in
  Linux*)   OS=Linux ;;
  Darwin*)  OS=Darwin ;;
  *)        red "Unsupported OS: $uname_s"; exit 1 ;;
esac

case "$uname_m" in
  x86_64 | amd64)  ARCH=amd64 ;;
  arm64 | aarch64) ARCH=arm64 ;;
  *)               red "Unsupported arch: $uname_m"; exit 1 ;;
esac

# Resolve latest version from GitHub API.
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep -m1 '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${VERSION:-}" ]; then
  red "Could not resolve the latest release from GitHub."
  exit 1
fi

ASSET="${BIN}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

green "Installing ${BIN} ${VERSION} for ${OS}/${ARCH}..."
yellow "Source: ${URL}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/${ASSET}"
tar -xzf "$TMP/${ASSET}" -C "$TMP"

# Pick install directory.
if [ -w "/usr/local/bin" ]; then
  DEST=/usr/local/bin
elif command -v sudo >/dev/null 2>&1; then
  DEST=/usr/local/bin
  SUDO=sudo
else
  DEST="$HOME/.local/bin"
  mkdir -p "$DEST"
fi

${SUDO:-} install -m 0755 "$TMP/${BIN}" "$DEST/${BIN}"

green "Installed ${DEST}/${BIN}"

if ! command -v "$BIN" >/dev/null 2>&1; then
  yellow "Note: $DEST is not on your PATH yet."
  yellow "Add this line to your shell rc (~/.bashrc, ~/.zshrc):"
  yellow "    export PATH=\"$DEST:\$PATH\""
fi

green "Run it with:  ${BIN}"
