#!/usr/bin/env bash

set -e
set -o pipefail
set -u

echo "generating public schema"
go run github.com/swaggo/swag/cmd/swag \
  init \
  --parseDependency \
  --output docs/public \
  --parseInternal -g main.go \
  --markdownFiles docs/public/descriptions \
  -g main.go \
  -t apps,components,installs,installers,general,orgs,releases,sandboxes,vcs,runners

echo "validating public schema"
go run github.com/go-swagger/go-swagger/cmd/swagger \
  validate ./docs/public/swagger.json

echo "generating admin schema"
go run github.com/swaggo/swag/cmd/swag \
  init \
  --instanceName admin \
  --output docs/admin \
  --parseDependency \
  --parseInternal -g main.go \
  --markdownFiles docs/admin/descriptions \
  -t orgs/admin,apps/admin,general/admin,sandboxes/admin,installs/admin,installers/admin,components/admin,runners/admin

echo "validating admin schema"
go run github.com/go-swagger/go-swagger/cmd/swagger \
  validate ./docs/admin/admin_swagger.json

echo "generating runner schema"
go run github.com/swaggo/swag/cmd/swag \
  init \
  --instanceName runner \
  --output docs/runner \
  --parseDependency \
  --parseInternal -g main.go \
  --markdownFiles docs/runner/descriptions \
  -t orgs/runner,apps/runner,general/runner,sandboxes/runner,installs/runner,installers/runner,components/runner,runners/runner

echo "validating admin schema"
go run github.com/go-swagger/go-swagger/cmd/swagger \
  validate ./docs/runner/runner_swagger.json
