#!/usr/bin/env bash

set -e
set -o pipefail
set -u

echo "generating public schema"
go run github.com/swaggo/swag/cmd/swag \
  init \
  --parseDependency \
  --parseInternal -g main.go \
  --markdownFiles docs/descriptions \
  -g main.go \
  -t apps,components,installs,installers,general,orgs,releases,sandboxes,vcs

echo "validating public schema"
go run github.com/go-swagger/go-swagger/cmd/swagger \
  validate ./docs/swagger.json

echo "generating admin schema"
go run github.com/swaggo/swag/cmd/swag \
  init \
  --instanceName admin \
  --output admin \
  --parseDependency \
  --parseInternal -g main.go \
  --markdownFiles admin/descriptions \
  -t orgs/admin,orgs/admin,apps/admin,general/admin,sandboxes/admin,installs/admin,installers/admin,components/admin

echo "validating admin schema"
go run github.com/go-swagger/go-swagger/cmd/swagger \
  validate ./admin/admin_swagger.json
