#!/usr/bin/env bash


set -e
set -o pipefail
set -u


function validate_public_schema() {
  echo "validating public schema"
  go run github.com/go-swagger/go-swagger/cmd/swagger \
    validate ./docs/public/swagger.json
}

function gen_public_schema() {
  echo "generating public schema"
  go run github.com/swaggo/swag/cmd/swag \
    init \
    --parseDependency \
    --output docs/public \
    --parseInternal \
    -g public.go \
    --markdownFiles docs/public/descriptions \
    -t apps,components,installs,installers,general,orgs,releases,sandboxes,vcs,runners
  # sleep to avoid confusingly interpolated validation & runner logs
  sleep .25
}

function public_schema() {
  gen_public_schema      > >(sed 's/^/[public] /') 2> >(sed 's/^/[public][error] /' >&2)
  validate_public_schema > >(sed 's/^/[public] /') 2> >(sed 's/^/[public][error] /' >&2)
}

function validate_admin_schema(){
  echo "validating admin schema"
  go run github.com/go-swagger/go-swagger/cmd/swagger \
    validate ./docs/admin/admin_swagger.json
}

function gen_admin_schema() {
  echo "generating admin schema"
  go run github.com/swaggo/swag/cmd/swag \
    init \
    --instanceName admin \
    --output docs/admin \
    --parseDependency \
    --parseInternal -g admin.go \
    --markdownFiles docs/admin/descriptions \
    -t orgs/admin,apps/admin,general/admin,sandboxes/admin,installs/admin,installers/admin,components/admin,runners/admin
  # sleep to avoid confusingly interpolated validation & runner logs
  sleep .25
}

function admin_schema() {
  gen_admin_schema       > >(sed 's/^/[admin] /') 2> >(sed 's/^/[admin][error] /' >&2)
  validate_admin_schema  > >(sed 's/^/[admin] /') 2> >(sed 's/^/[admin][error] /' >&2)
}


function validating_runner_schema(){
  echo "🏃 validating runner schema"
  go run github.com/go-swagger/go-swagger/cmd/swagger \
    validate ./docs/runner/runner_swagger.json
}

function gen_runner_schema() {
  echo "generating runner schema"
  go run github.com/swaggo/swag/cmd/swag \
    init \
    --instanceName runner \
    --output docs/runner \
    --parseDependency \
    --parseInternal -g runner.go \
    --markdownFiles docs/runner/descriptions \
    -t orgs/runner,apps/runner,general/runner,sandboxes/runner,installs/runner,installers/runner,components/runner,runners/runner
  # sleep to avoid confusingly interpolated validation & runner logs
  sleep .25
}

function runner_schema() {
  gen_runner_schema        > >(sed 's/^/[runner] /') 2> >(sed 's/^/[runner][error] /' >&2)
  validating_runner_schema > >(sed 's/^/[runner] /') 2> >(sed 's/^/[runner][error] /' >&2)
}


function generate() {
  echo
  echo " 🙈 Generating"
  echo

  public_schema &\
  admin_schema  &\
  runner_schema &\
  wait

  echo
  echo " ✅ Generated"
  echo

  exit 0
}

generate
