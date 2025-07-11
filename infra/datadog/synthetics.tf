resource "datadog_synthetics_test" "main" {
  # Only create resources if synthetics configuration exists
  for_each = try(local.vars.synthetics, {})

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
    monitor_priority = each.value.options_list.monitor_priority
  }
}

resource "datadog_synthetics_test" "installers" {
  # Only create resources if installer_synthetics configuration exists
  for_each = try(local.vars.installer_synthetics.installers, {})

  name      = "[${var.env}] Installer Uptime - ${each.key}"
  type      = try(local.vars.installer_synthetics.config.type, "api")
  subtype   = try(local.vars.installer_synthetics.config.subtype, "http")
  status    = try(local.vars.installer_synthetics.config.status, "paused")
  message   = try(local.vars.installer_synthetics.config.message, "Installer uptime check")
  locations = try(local.vars.installer_synthetics.config.locations, ["aws:us-east-1"])
  tags      = try(local.vars.installer_synthetics.config.tags, [])

  request_definition {
    method = try(local.vars.installer_synthetics.config.request_definition.method, "GET")
    url    = each.value
  }

  request_headers = try(local.vars.installer_synthetics.config.request_headers, {})

  assertion {
    type     = try(local.vars.installer_synthetics.config.assertion.type, "statusCode")
    operator = try(local.vars.installer_synthetics.config.assertion.operator, "is")
    target   = try(local.vars.installer_synthetics.config.assertion.target, "200")
  }

  options_list {
    tick_every = try(local.vars.installer_synthetics.config.options_list.tick_every, 900)
    retry {
      count    = try(local.vars.installer_synthetics.config.options_list.retry.count, 2)
      interval = try(local.vars.installer_synthetics.config.options_list.retry.interval, 300)
    }
    monitor_options {
      renotify_interval = try(local.vars.installer_synthetics.config.options_list.monitor_options.renotify_interval, 120)
    }
    monitor_priority = try(local.vars.installer_synthetics.config.options_list.monitor_priority, 3)
  }
}
