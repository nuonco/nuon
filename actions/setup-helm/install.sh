#!/usr/bin/env bash

[[ -n "${TRACE:-}" ]] && set -x

echo "::group::Install helm plugins"
trap 'echo "::endgroup::"' EXIT

HELM_S3="s3"

declare -A plugins
plugins[$HELM_S3]=https://github.com/hypnoglow/helm-s3.git

declare -A versions
versions[$HELM_S3]="v0.13.0"

for plugin in "${!plugins[@]}";
do
    if helm plugin list | grep -q "$plugin";
    then
        echo "Skipping plugin: $plugin. Already installed"
        continue;
    fi

    echo "Installing plugin: $plugin"
    # helm plugin install "${plugins[$plugin]}" --version "${versions[$plugin]}" &> /dev/null || exit 1
    helm plugin install "${plugins[$plugin]}" --version "${versions[$plugin]}" || exit 1
done

exit 0
