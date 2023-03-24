# the grafana cloud provider is used to manage grafana cloud - create stacks, keys etc.
resource "grafana_cloud_api_key" "otel" {
  provider = grafana.org

  cloud_org_slug = data.grafana_cloud_organization.current.slug
  name           = "nuon-main-otel"
  role           = "MetricsPublisher"
}

# the grafana provider is used for creating things within the "grafana" instance - such as alerts, dashboards etc.
provider "grafana" {
  url  = grafana_cloud_stack.main.url
  auth = grafana_api_key.main.key
}
