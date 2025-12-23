#!/usr/bin/env bash

set -e
set -o pipefail
set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SWAGGER_FILE="${SCRIPT_DIR}/../../services/ctl-api/docs/public/swagger.json"

if [ ! -f "$SWAGGER_FILE" ]; then
  echo >&2 "swagger.json not found at $SWAGGER_FILE"
  echo >&2 "run 'go run cmd/gen/main.go' from services/ctl-api first"
  exit 1
fi

echo >&2 "generating with OAPI spec from $SWAGGER_FILE"
go run github.com/go-swagger/go-swagger/cmd/swagger@v0.33.0 \
  generate \
  client \
  --skip-tag-packages \
  -f "$SWAGGER_FILE"

echo >&2 "generating mocks"
go run github.com/golang/mock/mockgen \
  -destination=mock.go \
  -source=client.go \
  -package=nuon
