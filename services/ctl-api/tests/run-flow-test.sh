#!/usr/bin/env bash
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ ${SHARDS:-1} -gt 1 && $# -eq 0 ]]; then
  mapfile -t tests < <(grep -hoE 'func \(e \*FlowTestSuite\) Test[A-Za-z0-9_]+' \
    "$here/../internal/pkg/flow/testworker/"*_test.go | awk '{print $4}' | sort)
  declare -a groups
  for i in "${!tests[@]}"; do
    idx=$((i % SHARDS))
    groups[idx]="${groups[idx]:+${groups[idx]}|}${tests[$i]}"
  done
  pids=()
  for i in $(seq 0 $((SHARDS - 1))); do
    SHARDS= PARALLEL=1 TESTWORKER_NAMESPACE="flowtest-$$-$i" \
      "$0" "^(${groups[$i]})$" > "/tmp/flow-shard-$i.log" 2>&1 &
    pids+=("$!")
  done
  rc=0
  for i in "${!pids[@]}"; do
    if wait "${pids[$i]}"; then
      echo "shard $i ok (/tmp/flow-shard-$i.log)"
    else
      rc=1
      echo "shard $i FAILED (/tmp/flow-shard-$i.log)"
    fi
  done
  exit $rc
fi

if [[ -n ${PARALLEL:-} ]]; then
  export TESTWORKER_NAMESPACE="${TESTWORKER_NAMESPACE:-flowtest-$$}"
fi
if [[ -n ${TESTWORKER_NAMESPACE:-} && $TESTWORKER_NAMESPACE != default ]]; then
  temporal operator namespace create --namespace "$TESTWORKER_NAMESPACE" \
    --address "${TEMPORAL_HOST:-localhost:7233}" >/dev/null 2>&1 || true
  until temporal operator namespace describe --namespace "$TESTWORKER_NAMESPACE" \
    --address "${TEMPORAL_HOST:-localhost:7233}" >/dev/null 2>&1; do
    sleep 0.5
  done
fi

set -a
# shellcheck disable=SC1091
source "$here/integration.env"
set +a
if [[ ${GITHUB_APP_KEY:-} != *"PRIVATE KEY"* ]]; then
  keyfile="${TMPDIR:-/tmp}/flow-testworker-gh-key.pem"
  [[ -s $keyfile ]] || openssl genrsa 2048 > "$keyfile" 2>/dev/null
  GITHUB_APP_KEY="$(cat "$keyfile")"
  export GITHUB_APP_KEY
fi

run='TestSuite'
if [[ $# -ge 1 && -n $1 ]]; then
  run="TestSuite/$1"
fi

args=(-run "$run" -timeout "${TIMEOUT:-45m}" -v -parallel "${NUON_FLOW_PARALLELISM:-1}")
if [[ -n ${SKIP:-} ]]; then
  args+=(-skip "$SKIP")
fi

exec go test "$here/../internal/pkg/flow/testworker/" "${args[@]}"
