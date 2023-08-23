#!/bin/bash

if [ "$NUON_DEBUG" = "true" ]
then
  set -x
fi

set -e
set -o pipefail
set -u

NAME=terraform-provider-nuon
BASE_URL=https://nuon-artifacts.s3.us-west-2.amazonaws.com/$NAME
VERSION=v0.0.1

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

DIR=~/.terraform.d/plugins/terraform.local/local/nuon/0.0.1/${OS}_${ARCH}
read -ep "Installing nuon terraform provider into $DIR, would you like to proceed? " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 1
fi

echo "calculating latest version..."
VERSION=$(curl -s $BASE_URL/latest.txt)
echo "✅ using version ${VERSION}..."

echo "ensuring directories exist"
mkdir -p $DIR

echo "fetching binary for ${OS} ${ARCH}..."
curl -s -o $DIR/${NAME}_${VERSION} $BASE_URL/$VERSION/${NAME}_${OS}_${ARCH}
echo "✅ fetching binary for ${OS} ${ARCH}..."

echo "making binary executable..."
chmod +x $DIR/${NAME}_${VERSION}
echo "✅ nuon should be ready to use"

echo "🚀 To get started, please run - nuon login"
