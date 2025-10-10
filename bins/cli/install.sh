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

echo "fetching binary for ${OS} ${ARCH}..."
curl -s -o $DIR/$NAME $BASE_URL/$VERSION/${NAME}_${OS}_${ARCH}
echo "✅ fetching binary for ${OS} ${ARCH}..."

echo "making binary executable..."
chmod +x $DIR/$NAME
echo "✅ nuon should be ready to use"

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
