#!/usr/bin/env bats

@test "accept because it does not read secrets" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/role-not-reading-secret.json

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request accepted
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*true') -ne 0 ]
}

@test "reject because it reads secrets" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/role-reading-secret.json

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request rejected
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : '.*Role reading secrets is not allowed.*') -ne 0 ]
}

@test "reject clusterrole because it reads secrets" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/clusterrole-reading-secret.json

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request rejected
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : '.*ClusterRole reading secrets is not allowed.*') -ne 0 ]
}

@test "reject role with wildcard resources" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/role-wildcard-resources.json

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request rejected
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : '.*Role reading secrets is not allowed.*') -ne 0 ]
}

@test "reject role with wildcard verbs" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/role-wildcard-verbs.json

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request rejected
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : '.*Role reading secrets is not allowed.*') -ne 0 ]
}

@test "reject role with mixed permissions (pods and secrets)" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/role-mixed-permissions.json

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request rejected
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : '.*Role reading secrets is not allowed.*') -ne 0 ]
}
