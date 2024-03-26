resource "datadog_monitor_json" "kubernetes_container_waiting" {
  for_each = { for monitor in local.vars.monitors : monitor.json => monitor }

  monitor = each.value.json
}
