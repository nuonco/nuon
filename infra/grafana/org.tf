data "grafana_cloud_organization" "current" {
  provider = grafana.org
  slug     = "nuon"
}
