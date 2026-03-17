#!/bin/bash

if [ "$NUON_DEBUG" = "true" ]
then
  set -x
fi

set -e
set -o pipefail
set -u

NO_INPUT=false
while [[ $# -gt 0 ]]; do
  case $1 in
    --no-input)
      NO_INPUT=true
      shift
      ;;
    *)
      shift
      ;;
  esac
done

BASE_URL=https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli
LSP_BASE_URL=https://nuon-artifacts.s3.us-west-2.amazonaws.com/lsp
# Create a temporary directory for downloading the binaries
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
if [ "$NO_INPUT" = false ]; then
  read -ep "press \"y\" to proceed: " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]
  then
      exit 1
  fi
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

# Function to download and install a binary
# Args: $1=binary_name (e.g., "nuon" or "nuon-lsp"), $2=base_url, $3=optional (true/false)
download_and_install_binary() {
  local NAME=$1
  local URL=${2:-$BASE_URL}
  local OPTIONAL=${3:-false}
  local success=false

  # Try gzip compressed binary first
  echo "fetching compressed binary for ${OS} ${ARCH}..."
  compressed_url="$URL/$VERSION/${NAME}_${OS}_${ARCH}.gz"

  set +e
  http_response=$(curl -s -f -w "%{http_code}" -o "$TEMP_DIR/$NAME.gz" "$compressed_url" 2>/dev/null)
  status=$?
  set -e

  if [ $status -eq 0 ] && [ "$http_response" = "200" ]; then
    echo "✅ compressed binary downloaded, extracting..."

    if gunzip -f "$TEMP_DIR/$NAME.gz" &> /dev/null; then
      echo "✅ extraction successful"

      if [ -f "$TEMP_DIR/$NAME" ]; then
        echo "moving binary to $DIR/$NAME..."
        mv "$TEMP_DIR/$NAME" "$DIR/$NAME"
        echo "making binary executable..."
        chmod +x "$DIR/$NAME"
        echo "✅ $NAME should be ready to use"
        success=true
      else
        echo "⚠️  extraction succeeded but binary not found, falling back..."
      fi
    else
      echo "⚠️  extraction failed, falling back to uncompressed binary..."
    fi
  else
    if [ "$OPTIONAL" = "false" ]; then
      echo "⚠️  compressed binary not available (HTTP status: $http_response), falling back..."
    fi
  fi

  # Cleanup failed attempt
  rm -f "$TEMP_DIR/$NAME.gz"

  # Fallback to uncompressed binary
  if [ "$success" = false ]; then
    echo "fetching binary for ${OS} ${ARCH}..."

    # Check if binary exists before downloading
    set +e
    http_response=$(curl -s -f -w "%{http_code}" -o "$TEMP_DIR/$NAME" "$URL/$VERSION/${NAME}_${OS}_${ARCH}" 2>/dev/null)
    status=$?
    set -e

    if [ $status -eq 0 ] && [ "$http_response" = "200" ]; then
      echo "✅ fetching binary for ${OS} ${ARCH}..."

      echo "moving binary to $DIR/$NAME..."
      mv "$TEMP_DIR/$NAME" "$DIR/$NAME"
      echo "making binary executable..."
      chmod +x "$DIR/$NAME"
      echo "✅ $NAME should be ready to use"
      success=true
    else
      if [ "$OPTIONAL" = "true" ]; then
        # Silently skip optional binaries that don't exist
        rm -f "$TEMP_DIR/$NAME"
        return 0
      else
        echo "❌ failed to download $NAME (HTTP status: $http_response)"
        return 1
      fi
    fi
  fi
}

# Install nuon CLI
download_and_install_binary "nuon" "$BASE_URL" false

# Install nuon-lsp (Language Server) - optional, silently skips if not available
download_and_install_binary "nuon-lsp" "$LSP_BASE_URL" true

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
version=$(nuon version -j | jq -r .version)
if [ "$version" != "$VERSION" ]; then
  echo "nuon version returned $version, expected $VERSION. This usually means nuon was installed to a separate location outside of this script"
  exit 1
fi

# Cleanup temp directory
rm -rf "$TEMP_DIR"

echo "🚀 To get started, please run - nuon login"
