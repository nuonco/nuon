resource "datadog_monitor_json" "main" {
  for_each = local.vars.monitors

  monitor = each.value
}
