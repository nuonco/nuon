#!/usr/bin/env bash

set -e
set -o pipefail
set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SWAGGER_FILE="${SCRIPT_DIR}/../../services/ctl-api/docs/stack/stack_swagger.json"
STAMP_FILE="${SCRIPT_DIR}/.swagger.stamp"

if [ ! -f "$SWAGGER_FILE" ]; then
  echo >&2 "stack_swagger.json not found at $SWAGGER_FILE"
  echo >&2 "attempting to generate specs via services/ctl-api/cmd/gen"
  (cd "$REPO_DIR/services/ctl-api" && go run cmd/gen/main.go --targets stack)
fi

if [ ! -f "$SWAGGER_FILE" ]; then
  echo >&2 "stack_swagger.json still missing after generation attempt"
  exit 1
fi

swagger_hash="$(shasum -a 256 "$SWAGGER_FILE" | awk '{print $1}')"
previous_hash=""
if [ -f "$STAMP_FILE" ]; then
  previous_hash="$(cat "$STAMP_FILE")"
fi

if [ "$swagger_hash" != "$previous_hash" ]; then
  echo >&2 "generating with OAPI spec from $SWAGGER_FILE"
  rm -rf models/ client/
  go run github.com/go-swagger/go-swagger/cmd/swagger@v0.33.0 \
    generate \
    client \
    --skip-tag-packages \
    -f "$SWAGGER_FILE"
  echo "$swagger_hash" > "$STAMP_FILE"
else
  echo >&2 "swagger spec unchanged, skipping client generation"
fi

# No mockgen step, unlike nuon-go/nuon-runner-go: mock.go is gitignored repo-wide,
# so the golang/mock require would reach the terraform provider's module graph to
# support a file that is never committed. Client has two methods; hand-fake it.
