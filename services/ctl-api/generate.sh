#!/usr/bin/env bash

echo "generating public schema"
go run github.com/swaggo/swag/cmd/swag \
  init \
  --parseGoList \
  -t orgs,apps,installs,general,vcs,components,sandboxes,releases

echo "validating public schema"
go run github.com/go-swagger/go-swagger/cmd/swagger \
  validate ./docs/swagger.json

echo "generating admin schema"
go run github.com/swaggo/swag/cmd/swag \
  init \
  --instanceName admin \
  --parseGoList \
  --output admin \
  --parseGoList \
  -t orgs/admin,orgs/admin,apps/admin,general/admin,sandboxes/admin,installs/admin

echo "validating admin schema"
go run github.com/go-swagger/go-swagger/cmd/swagger \
  validate ./admin/admin_swagger.json
