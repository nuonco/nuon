#!/usr/bin/env bash
# Run flow testworker integration tests from the repo root.
#
#   ./services/ctl-api/tests/run-flow-test.sh TestApprovalApproveContinues
#   ./services/ctl-api/tests/run-flow-test.sh 'TestFail.*'          # regex ok
#   ./services/ctl-api/tests/run-flow-test.sh                       # whole suite
#   SKIP='TestSuite/TestPin' ./services/ctl-api/tests/run-flow-test.sh
#
# Parallel shards: two runs sharing a namespace stomp each other (shared task
# queue + boot-time stale-workflow cleanup). PARALLEL=1 gives the run its own
# Temporal namespace so shards are fully isolated:
#
#   PARALLEL=1 ./services/ctl-api/tests/run-flow-test.sh 'TestApproval.*' &
#   PARALLEL=1 ./services/ctl-api/tests/run-flow-test.sh 'TestFail.*' &
#   wait
#
# SHARDS=4 splits the whole suite round-robin into that many concurrent
# namespace-isolated runs (logs land in /tmp/flow-shard-N.log):
#
#   SHARDS=4 ./services/ctl-api/tests/run-flow-test.sh
#
# Requires the local dev containers (postgres, clickhouse, temporal) running,
# the temporal `default` namespace, and the ctl_api_test databases created once
# with `go run ./cmd/nuontest` from services/ctl-api.
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
# The github client is built from GITHUB_APP_KEY and needs a parseable PEM; the
# shared integration.env only carries a non-empty dummy, so swap in a throwaway.
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

args=(-run "$run" -timeout "${TIMEOUT:-45m}" -v)
if [[ -n ${SKIP:-} ]]; then
  args+=(-skip "$SKIP")
fi

exec go test "$here/../internal/pkg/flow/testworker/" "${args[@]}"
