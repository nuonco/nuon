# Declaring the first provider to be only used for creating the cloud-stack
provider "grafana" {
  alias = "org"

  cloud_api_key = var.grafana_cloud_api_key
}
