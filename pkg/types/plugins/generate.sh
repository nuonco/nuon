#!/usr/bin/env bash

set -e
set -o pipefail

if [[ -z "${ENABLE_BUF_BUILD}" ]]; then
  echo "skipping generating unless ENABLE_BUF_BUILD=true is set"
  exit 0
fi

buf generate
buf generate --template buf.gen.tag.yaml
