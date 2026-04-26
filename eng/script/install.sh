#!/usr/bin/env bash
set -euo pipefail

REPO="voidbear-io/akv"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux) OS_NAME="linux" ;;
  Darwin) OS_NAME="darwin" ;;
  CYGWIN*|MINGW32*|MSYS*|MINGW*|Windows_NT) OS_NAME="windows" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH_NAME="amd64" ;;
  arm64|aarch64) ARCH_NAME="arm64" ;;
  i386|i686) ARCH_NAME="386" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if [ -z "${AKV_INSTALL_DIR:-}" ]; then
  if [ "$OS_NAME" = "windows" ]; then
    AKV_INSTALL_DIR="$USERPROFILE/AppData/Local/Programs/bin"
  else
    AKV_INSTALL_DIR="$HOME/.local/bin"
  fi
fi

mkdir -p "$AKV_INSTALL_DIR"

EXT="tar.gz"
if [ "$OS_NAME" = "windows" ]; then
  EXT="zip"
fi

echo "Fetching latest release information for $REPO..."
if command -v jq >/dev/null 2>&1; then
  DOWNLOAD_URL=$(curl -fsSL "$API_URL" | jq -r ".assets[]? | select(.name | contains(\"akv-${OS_NAME}-${ARCH_NAME}-v\")) | select(.name | endswith(\".${EXT}\")) | .browser_download_url")
else
  DOWNLOAD_URL=$(curl -fsSL "$API_URL" | grep -o "https://github.com/.*/releases/download/.*/akv-${OS_NAME}-${ARCH_NAME}-v.*\.${EXT}")
fi

if [ -z "$DOWNLOAD_URL" ]; then
  echo "Error: Could not find a release for $OS_NAME $ARCH_NAME"
  exit 1
fi

TMP_DIR=$(mktemp -d)
TMP_FILE="$TMP_DIR/akv.$EXT"

curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"

if [ "$EXT" = "zip" ]; then
  unzip -q "$TMP_FILE" -d "$TMP_DIR"
else
  tar -xzf "$TMP_FILE" -C "$TMP_DIR"
fi

if [ "$OS_NAME" = "windows" ]; then
  mv "$TMP_DIR/akv.exe" "$AKV_INSTALL_DIR/"
  echo "akv installed to $AKV_INSTALL_DIR/akv.exe"
else
  mv "$TMP_DIR/akv" "$AKV_INSTALL_DIR/"
  chmod +x "$AKV_INSTALL_DIR/akv"
  echo "akv installed to $AKV_INSTALL_DIR/akv"
fi

rm -rf "$TMP_DIR"

echo "Installation complete! Run 'akv --help' to get started."
