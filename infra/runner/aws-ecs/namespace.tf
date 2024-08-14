resource "aws_service_discovery_http_namespace" "this" {
  name        = var.runner_id
  description = "CloudMap namespace for ${var.runner_id}"
}
