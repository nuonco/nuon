resource "datadog_monitor_json" "kubernetes_container_waiting" {
  for_each = { for monitor in local.values.datadog.monitors : monitor.json => monitor }

  monitor = each.value.json
}
