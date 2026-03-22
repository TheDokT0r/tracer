#!/bin/sh
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
URL="https://github.com/TheDokT0r/tracer/releases/latest/download/tracer-${OS}-${ARCH}.tar.gz"

curl -fsSL "$URL" | tar -xz -C /usr/local/bin
echo "tracer installed to /usr/local/bin/tracer"
