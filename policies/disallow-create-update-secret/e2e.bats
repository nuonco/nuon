#!/usr/bin/env bats

@test "accept because not a Secret" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/pod.json

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request accepted
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*true') -ne 0 ]
}

@test "reject because Secrets are not allowed" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/secret.json

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request rejected
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : '.*Creating or Updating Secrets is not allowed.*') -ne 0 ]
}
