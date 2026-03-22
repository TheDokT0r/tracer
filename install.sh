#!/bin/sh
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
URL="https://github.com/TheDokT0r/tracer/releases/latest/download/tracer-${OS}-${ARCH}.tar.gz"

INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

curl -fsSL "$URL" | tar -xz -C "$INSTALL_DIR"

echo "tracer installed to ${INSTALL_DIR}/tracer"

# Check if install dir is in PATH
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "Add ${INSTALL_DIR} to your PATH if it isn't already" ;;
esac
