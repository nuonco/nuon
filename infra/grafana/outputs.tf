output "otel" {
  value = {
    stack                            = grafana_cloud_stack.main.name
    api_key                          = nonsensitive(grafana_cloud_api_key.otel.key)
    prometheus_user_id               = grafana_cloud_stack.main.prometheus_user_id
    prometheus_remote_write_endpoint = grafana_cloud_stack.main.prometheus_remote_write_endpoint
  }
}
