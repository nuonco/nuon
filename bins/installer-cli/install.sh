#!/bin/sh
# installer-cli bootstrap. Downloads the right binary for the host OS/arch,
# then execs it with whatever args followed `--` in the curl-pipe invocation:
#
#   curl -fsSL <BASE_URL>/install.sh | sh -s -- provision <CREATE_RUN_URL>
#
# Override the download host with INSTALLER_CLI_BASE_URL or pin a version with
# INSTALLER_CLI_VERSION.

if [ "${INSTALLER_CLI_DEBUG:-}" = "true" ]; then
  set -x
fi

set -eu

# DEFAULT_BASE_URL: the devserver rewrites this line at request time.
BASE_URL="${INSTALLER_CLI_BASE_URL:-https://install.nuon.co/installer-cli}"

NAME=installer-cli
DIR=$(mktemp -d)
trap 'rm -rf "$DIR"' EXIT

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH=amd64 ;;
  aarch64) ARCH=arm64 ;;
esac

VERSION="${INSTALLER_CLI_VERSION:-}"
if [ -z "$VERSION" ]; then
  VERSION=$(curl -fsSL "$BASE_URL/latest.txt")
fi

echo "installer-cli: downloading ${OS}/${ARCH} (${VERSION})..." >&2

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
    echo "installer-cli: failed to download from $raw_url" >&2
    exit 1
  fi
fi

# Verify SHA256 checksum. Skip with INSTALLER_CLI_SKIP_CHECKSUM=true.
if [ "${INSTALLER_CLI_SKIP_CHECKSUM:-}" != "true" ]; then
  sha_url="$BASE_URL/$VERSION/${NAME}_${OS}_${ARCH}.sha256"
  if ! curl -fsSL -o "$DIR/$NAME.sha256" "$sha_url"; then
    echo "installer-cli: failed to fetch checksum from $sha_url" >&2
    exit 1
  fi
  expected=$(awk '{print $1}' "$DIR/$NAME.sha256")
  if command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$DIR/$NAME" | awk '{print $1}')
  elif command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$DIR/$NAME" | awk '{print $1}')
  else
    echo "installer-cli: no shasum/sha256sum available; set INSTALLER_CLI_SKIP_CHECKSUM=true to bypass" >&2
    exit 1
  fi
  if [ "$expected" != "$actual" ]; then
    echo "installer-cli: checksum mismatch (expected $expected, got $actual)" >&2
    exit 1
  fi
fi

chmod +x "$DIR/$NAME"

exec "$DIR/$NAME" "$@"
