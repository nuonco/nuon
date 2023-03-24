#!/usr/bin/env bash

set -euo pipefail

trap 'rm -rf charts *.tgz; git reset --hard @' EXIT

function main() {
    [[ -n "${TRACE:-}" ]] && set -x

    local version
    version=$(yq '.version' Chart.yaml)

    sed -i 's/name: temporal/name: infra-temporal/' Chart.yaml
    helm dependency update
    helm package .
    helm push "infra-temporal-$version.tgz" oci://431927561584.dkr.ecr.us-west-2.amazonaws.com
}

main "$@"

