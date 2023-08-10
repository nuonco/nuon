#!/bin/bash

set -e
set -o pipefail
set -u

BASE_URL=https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli
NAME=nuon

DIR=$HOME/bins
if [ ! -d "$DIR" ]; then
  DIR=/usr/local/bin

  # fall back to /usr/local/bin
  if [ ! -d $DIR ]; then
    # fall back to /bin
    DIR=/bin
  fi
fi

read -ep "Installing nuon cli into $DIR, would you like to proceed? " -n 1 -r
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
OS=$(uname -s |  awk '{print tolower($0)}')
echo "✅ using version ${OS}_${ARCH}..."

echo "calculating latest version..."
VERSION=$(curl -s $BASE_URL/latest.txt)
echo "✅ using version ${VERSION}..."

echo "fetching binary for ${OS} ${ARCH}..."
curl -s -o $DIR/$NAME $BASE_URL/$VERSION/${NAME}_${OS}_${ARCH}
echo "✅ fetching binary for ${OS} ${ARCH}..."

echo "making binary executable..."
chmod +x $DIR/$NAME
echo "✅ nuon should be ready to use"

echo "🚀 To get started, please run - nuon login"
