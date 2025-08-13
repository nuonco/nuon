#!/bin/bash

if [ "$NUON_DEBUG" = "true" ]
then
  set -x
fi

set -e
set -o pipefail
set -u

BASE_URL=https://nuon-artifacts.s3.us-west-2.amazonaws.com/runner
NAME=runner
RUNNER_VERSION="${1:-latest}"
DIR="${2:-/usr/local/bin}"

# Function to fetch and install the binary
fetch_binary() {
  local dir=$1
  local version=$2
  local os=$3
  local arch=$4

  echo "fetching binary for ${os} ${arch}..."
  local url="$BASE_URL/$version/${NAME}_${os}_${arch}"

  # Use curl with -f flag to fail on server errors like 404
  # Also store HTTP status code for checking
  http_response=$(curl -s -f -w "%{http_code}" -o $dir/$NAME "$url" 2>/dev/null)
  local status=$?

  if [ $status -ne 0 ] || [ "$http_response" = "404" ]; then
    echo "❌ Error: Failed to download binary from $url (HTTP status: $http_response)"
    return 1
  fi

  echo "✅ fetching binary for ${os} ${arch}..."

  echo "making binary executable..."
  chmod +x $dir/$NAME
  echo "✅ runner should be ready to use"
  return 0
}

if [ ! -d "$DIR" ]; then
  DIR=/usr/local/bin

  # fall back to /usr/local/bin
  if [ ! -d $DIR ]; then
    # fall back to /bin
    DIR=/bin
  fi
fi

read -ep "Installing runner into $DIR, would you like to proceed? " -n 1 -r
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
if [ -n "${RUNNER_VERSION:-}" ]; then
  echo "⚠️  trying to use version RUNNER_VERSION=${RUNNER_VERSION}"
  if fetch_binary "$DIR" "$RUNNER_VERSION" "$OS" "$ARCH"; then
    echo "✅ Successfully installed specified version ${RUNNER_VERSION}"
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
