#!/usr/bin/env bash

set -e
set -o pipefail
set -u

curl -vvv -H "Authorization: Bearer $NUON_API_TOKEN" -X POST --data '{"name": "My Org"}' localhost:8081/v1/orgs
