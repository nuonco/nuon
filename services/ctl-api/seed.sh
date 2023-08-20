#!/usr/bin/env bash

BASE_URL=http://localhost:8081
INTERNAL_BASE_URL=http://localhost:8082
echo "a temporary seed script"

echo "creating sandbox..."
SANDBOX=$(curl -s  -H "Authorization: Bearer $NUON_API_TOKEN" -X POST --data '{"name":"aws-eks","description":"standard sandbox"}' $INTERNAL_BASE_URL/v1/sandboxes)
echo $SANDBOX
SANDBOX_ID=$(echo $SANDBOX | jq -r .id)

echo "creating sandbox release..."
SANDBOX_RELEASE=$(curl -s -H "Authorization: Bearer $NUON_API_TOKEN" -X POST --data '{"version":"08e7f11","terraform_version":"v1.5.3"}' "$INTERNAL_BASE_URL/v1/sandboxes/$SANDBOX_ID/release")
SANDBOX_RELEASE_ID=$(echo $SANDBOX_RELEASE | jq -r .id)
