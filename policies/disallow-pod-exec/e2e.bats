#!/usr/bin/env bats

@test "reject because it is a pod exec" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/pod_exec.json

  echo "output = ${output}"

  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : '.*Exec into pods is not allowed.*') -ne 0 ]
}

@test "accept because it is not a pod exec" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/pod_create.json

  echo "output = ${output}"

  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*true') -ne 0 ]
}
