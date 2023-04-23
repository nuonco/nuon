output "test_string" {
  value = "test_string"
}

output "test_number" {
  value = 1
}

output "test_map" {
  value = { "number" : 1, "string" : "a" }
}

output "cluster_name" {
  value = "cluster-name"
}

output "cluster_endpoint" {
  value = "cluster-endpoint"
}

output "cluster_certificate_authority_data" {
  value = "cluster certificate authority data field. This field must be twenty characters at least"
}
