#!/bin/sh
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
URL="https://github.com/TheDokT0r/tracer/releases/latest/download/tracer-${OS}-${ARCH}.tar.gz"

INSTALL_DIR="/usr/local/bin"

if [ -w "$INSTALL_DIR" ]; then
  curl -fsSL "$URL" | tar -xz -C "$INSTALL_DIR"
else
  curl -fsSL "$URL" | sudo tar -xz -C "$INSTALL_DIR"
fi

echo "tracer installed to ${INSTALL_DIR}/tracer"
