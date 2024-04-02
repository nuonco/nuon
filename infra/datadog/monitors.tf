resource "datadog_monitor_json" "kubernetes_container_waiting" {
  for_each = local.vars.monitors

  monitor = each.value
}
