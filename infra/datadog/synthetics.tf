resource "datadog_synthetics_test" "main" {
  for_each = local.vars.synthetics

  name      = each.value.name
  type      = each.value.type
  subtype   = each.value.subtype
  status    = each.value.status
  message   = each.value.message
  locations = each.value.locations
  tags      = each.value.tags

  request_definition {
    method = each.value.request_definition.method
    url    = each.value.request_definition.url
  }

  request_headers = each.value.request_headers

  assertion {
    type     = each.value.assertion.type
    operator = each.value.assertion.operator
    target   = each.value.assertion.target
  }

  options_list {
    tick_every = each.value.options_list.tick_every
    retry {
      count    = each.value.options_list.retry.count
      interval = each.value.options_list.retry.interval
    }
    monitor_options {
      renotify_interval = each.value.options_list.monitor_options.renotify_interval
    }
  }
}
