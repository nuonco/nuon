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
    monitor_priority = each.value.options_list.monitor_priority
  }
}

resource "datadog_synthetics_test" "installers" {
  for_each = local.vars.installer_synthetics.installers

  name      = "[${var.env}] Installer Uptime - ${each.key}"
  type      = local.vars.installer_synthetics.config.type
  subtype   = local.vars.installer_synthetics.config.subtype
  status    = local.vars.installer_synthetics.config.status
  message   = local.vars.installer_synthetics.config.message
  locations = local.vars.installer_synthetics.config.locations
  tags      = local.vars.installer_synthetics.config.tags

  request_definition {
    method = local.vars.installer_synthetics.config.request_definition.method
    url    = "${each.value}"
  }

  request_headers = local.vars.installer_synthetics.config.request_headers

  assertion {
    type     = local.vars.installer_synthetics.config.assertion.type
    operator = local.vars.installer_synthetics.config.assertion.operator
    target   = local.vars.installer_synthetics.config.assertion.target
  }


  options_list {
    tick_every = local.vars.installer_synthetics.config.options_list.tick_every
    retry {
      count    = local.vars.installer_synthetics.config.options_list.retry.count
      interval = local.vars.installer_synthetics.config.options_list.retry.interval
    }
    monitor_options {
      renotify_interval = local.vars.installer_synthetics.config.options_list.monitor_options.renotify_interval
    }
    monitor_priority = local.vars.installer_synthetics.config.options_list.monitor_priority
  }
}
