#!/bin/sh
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
URL="https://github.com/orkwitzel/tracer/releases/latest/download/tracer-${OS}-${ARCH}.tar.gz"

INSTALL_DIR="$HOME/.local/bin"
MAN_DIR="$HOME/.local/share/man/man1"
mkdir -p "$INSTALL_DIR" "$MAN_DIR"

TMP=$(mktemp -d)
curl -fsSL "$URL" | tar -xz -C "$TMP"
mv "$TMP/tracer" "$INSTALL_DIR/tracer"
mv "$TMP/tracer.1" "$MAN_DIR/tracer.1"
chmod +x "$INSTALL_DIR/tracer"
rm -rf "$TMP"

echo "tracer installed to ${INSTALL_DIR}/tracer"
echo "man page installed to ${MAN_DIR}/tracer.1"

case ":$PATH:" in
  *":${INSTALL_DIR}:"*)
    ;;
  *)
    echo ""
    echo "${INSTALL_DIR} is not in your PATH."
    # Detect shell config file
    SHELL_NAME=$(basename "$SHELL" 2>/dev/null || echo "")
    case "$SHELL_NAME" in
      zsh)  SHELL_RC="$HOME/.zshrc" ;;
      bash) SHELL_RC="$HOME/.bashrc" ;;
      fish) SHELL_RC="$HOME/.config/fish/config.fish" ;;
      *)    SHELL_RC="" ;;
    esac
    EXPORT_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""
    if [ -n "$SHELL_RC" ]; then
      printf "Would you like to add it to %s? [Y/n] " "$SHELL_RC"
      read -r REPLY < /dev/tty
      case "$REPLY" in
        [nN]*)
          echo "Skipped. You can add it manually:"
          echo "  $EXPORT_LINE"
          ;;
        *)
          echo "" >> "$SHELL_RC"
          echo "# Added by tracer installer" >> "$SHELL_RC"
          echo "$EXPORT_LINE" >> "$SHELL_RC"
          echo "Added to ${SHELL_RC}. Run 'source ${SHELL_RC}' or restart your shell."
          ;;
      esac
    else
      echo "Add it manually to your shell config:"
      echo "  $EXPORT_LINE"
    fi
    ;;
esac

case ":$MANPATH:" in
  *":${MAN_DIR%/man1}:"*) ;;
  *) echo "Run 'export MANPATH=\"${MAN_DIR%/man1}:\$MANPATH\"' to enable 'man tracer'" ;;
esac
