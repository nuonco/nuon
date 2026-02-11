#!/usr/bin/env bash

set -e
set -o pipefail
set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SWAGGER_FILE="${SCRIPT_DIR}/../../services/ctl-api/docs/public/swagger.json"
STAMP_FILE="${SCRIPT_DIR}/.swagger.stamp"

if [ ! -f "$SWAGGER_FILE" ]; then
  echo >&2 "swagger.json not found at $SWAGGER_FILE"
  echo >&2 "attempting to generate specs via services/ctl-api/cmd/gen"
  (cd "$REPO_DIR/services/ctl-api" && go run cmd/gen/main.go --targets sdk)
fi

if [ ! -f "$SWAGGER_FILE" ]; then
  echo >&2 "swagger.json still missing after generation attempt"
  exit 1
fi

swagger_hash="$(shasum -a 256 "$SWAGGER_FILE" | awk '{print $1}')"
previous_hash=""
if [ -f "$STAMP_FILE" ]; then
  previous_hash="$(cat "$STAMP_FILE")"
fi

if [ "$swagger_hash" != "$previous_hash" ]; then
  echo >&2 "generating with OAPI spec from $SWAGGER_FILE"
  go run github.com/go-swagger/go-swagger/cmd/swagger@v0.33.0 \
    generate \
    client \
    --skip-tag-packages \
    -f "$SWAGGER_FILE"
  echo "$swagger_hash" > "$STAMP_FILE"
else
  echo >&2 "swagger spec unchanged, skipping client generation"
fi

echo >&2 "generating mocks"
go run github.com/golang/mock/mockgen@latest \
  -destination=mock.go \
  -source=client.go \
  -package=nuon
