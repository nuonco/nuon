#!/usr/bin/env bats

@test "accept because not a Role or ClusterRole" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/pod.json

  echo "output = ${output}"

  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*true') -ne 0 ]
}

@test "reject because Roles are not allowed" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/role.json

  echo "output = ${output}"

  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : '.*Creating Roles or ClusterRoles is not allowed.*') -ne 0 ]
}

@test "reject because ClusterRoles are not allowed" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/clusterrole.json

  echo "output = ${output}"

  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : '.*Creating Roles or ClusterRoles is not allowed.*') -ne 0 ]
}
