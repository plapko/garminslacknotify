#!/usr/bin/env bash
set -euo pipefail

REPO="plapko/garminslacknotify"
TARGET="/usr/local/bin/garminslacknotify"

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

ASSET_PATTERN="garminslacknotify_*_linux_${ARCH}.tar.gz"

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

echo "[1/4] Downloading latest release from $REPO (linux/$ARCH) ..."
gh release download -R "$REPO" -p "$ASSET_PATTERN" -D "$TMP_DIR" --clobber

ARCHIVE="$(ls "$TMP_DIR"/*.tar.gz 2>/dev/null | head -1)"
if [ -z "$ARCHIVE" ]; then
  echo "Archive not found in $TMP_DIR"
  exit 1
fi

echo "[2/4] Extracting archive ..."
tar -xzf "$ARCHIVE" -C "$TMP_DIR"

BIN="$TMP_DIR/garminslacknotify"
if [ ! -f "$BIN" ]; then
  echo "Binary not found after extraction: $BIN"
  exit 1
fi

echo "[3/4] Installing binary to $TARGET ..."
chmod +x "$BIN"

if [ -w "$(dirname "$TARGET")" ]; then
  mv "$BIN" "$TARGET"
else
  sudo mv "$BIN" "$TARGET"
fi

echo "[4/4] Done"
echo "Installed: $TARGET"

"$TARGET" --version && echo "Binary check: OK" || echo "Binary check: command returned non-zero"
