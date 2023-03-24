#!/usr/bin/env bash

set -euo pipefail

function main() {
    [[ -n "${TRACE:-}" ]] && set -x

    local context
    context=$(kubectl config current-context)

    local resp

    read -r -p "Continue using $context kubectl context? [Y/n] " resp
    case "$resp" in
        n|N)
            exit 1
            ;;
        *)
            ;;
    esac

    local name
    read -r -p "Namespace name: " name

    local retention
    read -r -p "Retention (in days): " retention

    local description
    read -r -p "Description: " description

    read -r -p "Continue using name: $name? [Y/n] " resp
    case "$resp" in
        n|N)
            exit 1
            ;;
        *)
            ;;
    esac

    read -r -p "Continue using retention: $retention? [Y/n] " resp
    case "$resp" in
        n|N)
            exit 1
            ;;
        *)
            ;;
    esac

    read -r -p "Continue using description: $description? [Y/n] " resp
    case "$resp" in
        n|N)
            exit 1
            ;;
        *)
            ;;
    esac

    run "$name" "$retention" "$description"

    return
}

function run() {
    local name="$1"
    local retention="$2"
    local description="$3"

    kubectl \
        exec \
        -it \
        --namespace=temporal \
        deployment/temporal-admintools \
        -- tctl --namespace "$name" namespace register \
            --description "$description" \
            --retention "$retention"
}

main "$@"
