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

    local workspace
    workspace=$(terraform workspace show)

    read -r -p "Continue using $workspace terraform workspace? [Y/n] " resp
    case "$resp" in
        n|N)
            exit 1
            ;;
        *)
            ;;
    esac

    echo "...Fetching connection details from terraform outputs..."

    local output
    local address
    local port
    local user
    local pass

    output=$(terraform output -json)
    address=$(echo "$output" | jq -r '.db_instance_address.value')
    port=$(echo "$output" | jq -r '.db_instance_port.value')
    user=$(echo "$output" | jq -r '.db_instance_username.value')
    pass=$(echo "$output" | jq -r '.db_instance_password.value')
    version=$(echo "$output" | jq -r '.image_tag.value')


    echo "Dropping you into the admin tools image..."
    echo "Environment variables will be set for use with temporal-sql-tool..."
    run "$address" "$port" "$user" "$pass" "$version"

    return
}

function run() {
    local address="$1"
    local port="$2"
    local user="$3"
    local pass="$4"
    local version="$5"

    kubectl \
        run \
        -it \
        --rm \
        --namespace=temporal \
        "nuon-temporal-admintools-$(date +"%s")" \
        --image=temporalio/admin-tools:"$version" \
        --env="SQL_HOST=$address" \
        --env="SQL_PORT=$port" \
        --env="SQL_USER=$user" \
        --env="SQL_PASSWORD=$pass" \
        --env="SQL_PLUGIN=postgres" \
        --env="VERSION=$version" \
        --command \
        -- sh
}

main "$@"
