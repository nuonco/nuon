#!/usr/bin/env bats

@test "accept because not a LoadBalancer service" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/service-clusterip.json

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request accepted
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*true') -ne 0 ]
}

@test "reject because LoadBalancer services are not allowed" {
  run kwctl run -e gatekeeper annotated-policy.wasm -r test_data/service_loadbalancer.json

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request rejected
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : '.*Service of type LoadBalancer are not allowed.*') -ne 0 ]
}
