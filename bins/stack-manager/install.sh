#!/bin/sh
# stack-manager bootstrap. Downloads the right binary for the host OS/arch,
# then execs it with whatever args followed `--` in the curl-pipe invocation:
#
#   curl -fsSL <BASE_URL>/install.sh | sh -s -- provision <CREATE_RUN_URL>
#
# Override the download host with STACK_MANAGER_BASE_URL or pin a version with
# STACK_MANAGER_VERSION.

if [ "${STACK_MANAGER_DEBUG:-}" = "true" ]; then
  set -x
fi

set -eu

# DEFAULT_BASE_URL: the devserver rewrites this line at request time.
BASE_URL="${STACK_MANAGER_BASE_URL:-https://cdn.public.nuon.co/stack-manager}"

NAME=stack-manager

# Install dir defaults to ~/.local/bin (XDG-standard, often already on PATH).
# We can't use ~/.nuon/bin — the `nuon` CLI keeps its config at ~/.nuon as a
# regular file, so `mkdir -p ~/.nuon/bin` fails. Override with
# STACK_MANAGER_INSTALL_DIR if you want it somewhere else.
INSTALL_DIR="${STACK_MANAGER_INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$INSTALL_DIR"

DIR=$(mktemp -d)
trap 'rm -rf "$DIR"' EXIT

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH=amd64 ;;
  aarch64) ARCH=arm64 ;;
esac

VERSION="${STACK_MANAGER_VERSION:-}"
if [ -z "$VERSION" ]; then
  VERSION=$(curl -fsSL "$BASE_URL/latest.txt")
fi

echo "stack-manager: downloading ${OS}/${ARCH} (${VERSION})..." >&2

success=0

# Try gzipped first.
gz_url="$BASE_URL/$VERSION/${NAME}_${OS}_${ARCH}.gz"
if curl -fsSL -o "$DIR/$NAME.gz" "$gz_url" 2>/dev/null; then
  if gunzip -f "$DIR/$NAME.gz" 2>/dev/null && [ -f "$DIR/$NAME" ]; then
    success=1
  fi
fi

# Fall back to raw binary.
if [ "$success" -eq 0 ]; then
  raw_url="$BASE_URL/$VERSION/${NAME}_${OS}_${ARCH}"
  if ! curl -fsSL -o "$DIR/$NAME" "$raw_url"; then
    echo "stack-manager: failed to download from $raw_url" >&2
    exit 1
  fi
fi

# Verify SHA256 checksum against the published checksums.txt. Skip with
# STACK_MANAGER_SKIP_CHECKSUM=true.
if [ "${STACK_MANAGER_SKIP_CHECKSUM:-}" != "true" ]; then
  sums_url="$BASE_URL/$VERSION/checksums.txt"
  if ! curl -fsSL -o "$DIR/checksums.txt" "$sums_url"; then
    echo "stack-manager: failed to fetch $sums_url" >&2
    exit 1
  fi
  expected=$(awk -v f="${NAME}_${OS}_${ARCH}" '$2 == f {print $1; exit}' "$DIR/checksums.txt")
  if [ -z "$expected" ]; then
    echo "stack-manager: no checksum entry for ${NAME}_${OS}_${ARCH} in checksums.txt" >&2
    exit 1
  fi
  if command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$DIR/$NAME" | awk '{print $1}')
  elif command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$DIR/$NAME" | awk '{print $1}')
  else
    echo "stack-manager: no shasum/sha256sum available; set STACK_MANAGER_SKIP_CHECKSUM=true to bypass" >&2
    exit 1
  fi
  if [ "$expected" != "$actual" ]; then
    echo "stack-manager: checksum mismatch (expected $expected, got $actual)" >&2
    exit 1
  fi
fi

chmod +x "$DIR/$NAME"

# Move the verified binary to its persistent home so the user can re-invoke
# `stack-manager` without re-downloading.
mv "$DIR/$NAME" "$INSTALL_DIR/$NAME"
INSTALLED="$INSTALL_DIR/$NAME"

echo "stack-manager: installed to $INSTALLED" >&2
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "stack-manager: add $INSTALL_DIR to your PATH to invoke it directly." >&2 ;;
esac

exec "$INSTALLED" "$@"
