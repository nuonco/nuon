#!/usr/bin/env bash

set -e
set -o pipefail
set -u

echo "generating public schema"
go run github.com/swaggo/swag/cmd/swag \
  init \
  --parseDependency \
  --markdownFiles docs/descriptions \
  -g main.go \
  -t orgs,apps,installs,general,vcs,components,sandboxes,releases

echo "validating public schema"
go run github.com/go-swagger/go-swagger/cmd/swagger \
  validate ./docs/swagger.json

echo "generating admin schema"
go run github.com/swaggo/swag/cmd/swag \
  init \
  --instanceName admin \
  --output admin \
  --parseGoList \
  --markdownFiles admin/descriptions \
  -t orgs/admin,orgs/admin,apps/admin,general/admin,sandboxes/admin,installs/admin,components/admin

echo "validating admin schema"
go run github.com/go-swagger/go-swagger/cmd/swagger \
  validate ./admin/admin_swagger.json
