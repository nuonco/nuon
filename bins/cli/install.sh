#!/bin/bash

if [ "$NUON_DEBUG" = "true" ]
then
  set -x
fi

set -e
set -o pipefail
set -u

BASE_URL=https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli
NAME=nuon
# Create a temporary directory for downloading the binary
TEMP_DIR=$(mktemp -d)

DIR=~/bin
if [ ! -d "$DIR" ]; then
  DIR=/usr/local/bin

  # fall back to /usr/local/bin
  if [ ! -d $DIR ]; then
    # fall back to /bin
    DIR=/bin
  fi
fi

echo "Installing nuon cli into $DIR"
read -ep "press \"y\" to proceed: " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 1
fi

echo "checking OS and Architecture..."
set +e
dpkg_path=$(which dpkg)
set -e
if [ "$dpkg_path" = "" ]
then
  ARCH=$(uname -m)
else
  ARCH=$(dpkg --print-architecture)
fi

if [ "$ARCH" = "x86_64" ]; then
  ARCH=amd64
fi

OS=$(uname -s |  awk '{print tolower($0)}')
echo "✅ using version ${OS}_${ARCH}..."

# Check if version override is provided
if [ -n "${NUON_VERSION:-}" ]; then
  echo "⚠️  overriding version with NUON_VERSION=${NUON_VERSION}"
  VERSION=$NUON_VERSION
else
  echo "calculating latest version..."
  VERSION=$(curl -s $BASE_URL/latest.txt)
  echo "✅ using version ${VERSION}..."
fi

# Check if 7zip is available (from brew or system)
has_7z=false
if command -v 7z &> /dev/null || command -v 7za &> /dev/null || command -v 7zz &> /dev/null; then
  has_7z=true
  echo "7zip detected, will try compressed binary first..."
fi

# Try 7z compressed binary first if 7zip is available
if [ "$has_7z" = true ]; then
  echo "fetching compressed binary for ${OS} ${ARCH}..."
  compressed_url="$BASE_URL/$VERSION/${NAME}_${OS}_${ARCH}.7z"

  # Try to download .7z file
  http_response=$(curl -s -f -w "%{http_code}" -o "$TEMP_DIR/$NAME.7z" "$compressed_url" 2>/dev/null)
  status=$?

  if [ $status -eq 0 ] && [ "$http_response" = "200" ]; then
    echo "✅ compressed binary downloaded, extracting..."

    # Determine which 7z command to use (prefer 7zz, then 7z, then 7za)
    extract_cmd=""
    if command -v 7zz &> /dev/null; then
      extract_cmd="7zz"
    elif command -v 7z &> /dev/null; then
      extract_cmd="7z"
    elif command -v 7za &> /dev/null; then
      extract_cmd="7za"
    fi

    # Extract to temp directory
    if $extract_cmd x "$TEMP_DIR/$NAME.7z" -o"$TEMP_DIR" -y &> /dev/null; then
      echo "✅ extraction successful"

      # Move the binary
      if [ -f "$TEMP_DIR/${NAME}_${OS}_${ARCH}" ]; then
        echo "moving binary to $DIR/$NAME..."
        mv "$TEMP_DIR/${NAME}_${OS}_${ARCH}" "$DIR/$NAME"
        echo "making binary executable..."
        chmod +x "$DIR/$NAME"
        echo "✅ nuon should be ready to use"

        # Cleanup
        rm -f "$TEMP_DIR/$NAME.7z"
        # Skip fallback
        has_7z=false
      else
        echo "⚠️  extraction succeeded but binary not found, falling back..."
      fi
    else
      echo "⚠️  extraction failed, falling back to uncompressed binary..."
    fi
  else
    echo "⚠️  compressed binary not available (HTTP status: $http_response), falling back..."
  fi

  # Cleanup failed attempt
  rm -f "$TEMP_DIR/$NAME.7z"
fi

# Fallback to original logic (uncompressed binary) - only if 7z path didn't succeed
if [ "$has_7z" = true ] || [ ! -f "$DIR/$NAME" ]; then
  echo "fetching binary for ${OS} ${ARCH}..."
  curl -s -o "$TEMP_DIR/$NAME" "$BASE_URL/$VERSION/${NAME}_${OS}_${ARCH}"
  echo "✅ fetching binary for ${OS} ${ARCH}..."

  echo "moving binary to $DIR/$NAME..."
  mv "$TEMP_DIR/$NAME" "$DIR/$NAME"
  echo "making binary executable..."
  chmod +x "$DIR/$NAME"
  echo "✅ nuon should be ready to use"
fi

echo "ensuring installed correctly"
set +e
which nuon
which_status=$?
set -e
if [ $which_status -ne 0 ]; then
  echo "unable to find nuon, please make sure $DIR is on $PATH"
  exit 1
fi

echo "ensuring version is correct"
version=$(nuon version -j)
if [ "$version" != "$VERSION" ]; then
  echo "nuon version returned $version, expected $VERSION. This usually means nuon was installed to a separate location outside of this script"
  exit 1
fi

echo "🚀 To get started, please run - nuon login"
