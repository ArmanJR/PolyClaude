#!/bin/sh
set -e

REPO="ArmanJR/PolyClaude"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

VERSION=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d '"' -f4)
if [ -z "$VERSION" ]; then
  echo "Failed to fetch latest version" >&2
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${VERSION}/PolyClaude_${VERSION#v}_${OS}_${ARCH}.tar.gz"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading polyclaude ${VERSION} (${OS}/${ARCH})..."
curl -sSfL "$URL" -o "${TMPDIR}/polyclaude.tar.gz"
tar xzf "${TMPDIR}/polyclaude.tar.gz" -C "$TMPDIR"

if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/polyclaude" "${INSTALL_DIR}/polyclaude"
else
  sudo mv "${TMPDIR}/polyclaude" "${INSTALL_DIR}/polyclaude"
fi

echo "Installed polyclaude ${VERSION} to ${INSTALL_DIR}/polyclaude"
