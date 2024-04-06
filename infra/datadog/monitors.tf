resource "datadog_monitor_json" "main" {
  for_each = { for k, v in local.vars.monitors : k => v if v.enabled }

  monitor = each.value.json
}
