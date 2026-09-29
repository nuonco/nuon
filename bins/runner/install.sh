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
NO_INPUT=false

positional=()
for arg in "$@"; do
  case "$arg" in
    --no-input) NO_INPUT=true ;;
    *) positional+=("$arg") ;;
  esac
done

RUNNER_VERSION="${positional[0]:-latest}"
DIR="${positional[1]:-/usr/local/bin}"

fetch_binary() {
  local dir=$1
  local version=$2
  local os=$3
  local arch=$4

  echo "fetching binary for ${os} ${arch}..."
  local url="$BASE_URL/$version/${NAME}_${os}_${arch}"

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

  if [ ! -d $DIR ]; then
    DIR=/bin
  fi
fi

if [ "$NO_INPUT" = false ]; then
  read -ep "Installing runner into $DIR, would you like to proceed? " -n 1 -r
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

echo "calculating latest version..."
LATEST_VERSION=$(curl -s $BASE_URL/latest.txt)
echo "✅ latest version is ${LATEST_VERSION}"

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
  if fetch_binary "$DIR" "$LATEST_VERSION" "$OS" "$ARCH"; then
    echo "✅ Successfully installed latest version ${LATEST_VERSION}"
  else
    echo "❌ Failed to install latest version"
    exit 1
  fi
fi
