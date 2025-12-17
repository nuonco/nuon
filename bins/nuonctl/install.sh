#!/bin/bash

if [ "$NUON_DEBUG" = "true" ]
then
  set -x
fi

set -e
set -o pipefail
set -u

BASE_URL=https://nuon-artifacts.s3.us-west-2.amazonaws.com/nuonctl
NAME=nuonctl
# Create a temporary directory for downloading the binary
TEMP_DIR=$(mktemp -d)

# Function to fetch and install the binary
fetch_binary() {
  local dir=$1
  local version=$2
  local os=$3
  local arch=$4

  # Check if binary already exists
  if [ -f "$dir/$NAME" ]; then
    echo "checking existing binary version..."
    # Get current version, handle potential errors
    current_version=$($dir/$NAME version | tail -n 1)

    # Compare versions
    if [ "$current_version" = "$version" ]; then
      echo "✅ nuonctl version $version is already installed"
      return 0
    else
      echo "existing version: $current_version, requested version: $version"
    fi
  fi

  # Check if 7zip is available (from brew or system)
  local has_7z=false
  if command -v 7z &> /dev/null || command -v 7za &> /dev/null || command -v 7zz &> /dev/null; then
    has_7z=true
    echo "7zip detected, will try compressed binary first..."
  else
    echo "⚠️  7zip not found, falling back to uncompressed binary (slower download)"
    if command -v brew &> /dev/null; then
      echo "💡 Tip: Install 7zip for faster downloads: brew install sevenzip"
    fi
  fi

  # Try 7z compressed binary first if 7zip is available
  if [ "$has_7z" = true ]; then
    echo "fetching compressed binary for ${os} ${arch}..."
    local compressed_url="$BASE_URL/$version/${NAME}_${os}_${arch}.7z"

    # Try to download .7z file
    http_response=$(curl -s -f -w "%{http_code}" -o "$TEMP_DIR/$NAME.7z" "$compressed_url" 2>/dev/null)
    local status=$?

    if [ $status -eq 0 ] && [ "$http_response" = "200" ]; then
      echo "✅ compressed binary downloaded, extracting..."

      # Determine which 7z command to use (prefer 7zz, then 7z, then 7za)
      local extract_cmd=""
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
        if [ -f "$TEMP_DIR/${NAME}_${os}_${arch}" ]; then
          echo "moving binary to $dir..."
          mv "$TEMP_DIR/${NAME}_${os}_${arch}" "$dir/$NAME"
          echo "making binary executable..."
          chmod +x "$dir/$NAME"
          echo "✅ nuonctl should be ready to use"

          # Cleanup
          rm -f "$TEMP_DIR/$NAME.7z"
          return 0
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

  # Fallback to original logic (uncompressed binary)
  echo "fetching binary for ${os} ${arch}..."
  local url="$BASE_URL/$version/${NAME}_${os}_${arch}"

  # Use curl with -f flag to fail on server errors like 404
  # Also store HTTP status code for checking
  http_response=$(curl -s -f -w "%{http_code}" -o "$TEMP_DIR/$NAME" "$url" 2>/dev/null)
  local status=$?

  if [ $status -ne 0 ] || [ "$http_response" = "404" ]; then
    echo "❌ Error: Failed to download binary from $url (HTTP status: $http_response)"
    return 1
  fi

  echo "✅ fetching binary for ${os} ${arch}..."

  echo "moving binary to $dir..."
  mv "$TEMP_DIR/$NAME" "$dir/$NAME"
  echo "making binary executable..."
  chmod +x $dir/$NAME
  echo "✅ nuonctl should be ready to use"
  return 0
}

DIR=$HOME/bin
if [ ! -d "$DIR" ]; then
  DIR=/usr/local/bin

  # fall back to /usr/local/bin
  if [ ! -d $DIR ]; then
    # fall back to /bin
    DIR=/bin
  fi
fi

read -ep "Installing nuonctl into $DIR, would you like to proceed? " -n 1 -r
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

# Always fetch the latest version first
echo "calculating latest version..."
LATEST_VERSION=$(curl -s $BASE_URL/latest.txt)
echo "✅ latest version is ${LATEST_VERSION}"

# Try the provided version first, fall back to latest if it fails
if [ -n "${NUONCTL_VERSION:-}" ]; then
  echo "⚠️  trying to use version NUONCTL_VERSION=${NUONCTL_VERSION}"
  if fetch_binary "$DIR" "$NUONCTL_VERSION" "$OS" "$ARCH"; then
    echo "✅ Successfully installed specified version ${NUONCTL_VERSION}"
  else
    echo "⚠️  Specified version failed, falling back to latest version ${LATEST_VERSION}"
    if fetch_binary "$DIR" "$LATEST_VERSION" "$OS" "$ARCH"; then
      echo "✅ Successfully installed latest version ${LATEST_VERSION}"
    else
      echo "❌ Failed to install both specified and latest versions"
      exit 1
    fi
  fi
else
  # No specific version requested, use latest
  if fetch_binary "$DIR" "$LATEST_VERSION" "$OS" "$ARCH"; then
    echo "✅ Successfully installed latest version ${LATEST_VERSION}"
  else
    echo "❌ Failed to install latest version"
    exit 1
  fi
fi
