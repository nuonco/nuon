locals {
  # Check if byoc_synthetics exists in local.vars, if not use an empty map
  byoc_synthetics = try(local.vars.byoc_synthetics, {})
  
  # Only process instances if byoc_synthetics and its instances exist
  instances_map = try(
    {
      for item in flatten([
        for org_name, instances in local.byoc_synthetics.instances : [
          for instance_name, root_domain in instances : {
            key           = "${org_name}-${instance_name}"
            org_name      = org_name
            instance_name = instance_name
            root_domain   = root_domain
          }
        ]
      ]) : item.key => item
    },
    {} # Return empty map if there's an error (i.e., the required structure doesn't exist)
  )
}

resource "datadog_synthetics_test" "byoc-ctl-api" {
  # Only create resources if we have valid instances
  for_each = local.instances_map

  name      = "[byoc] ${each.value.org_name}: ${each.value.instance_name}: ctl-api"
  type      = try(local.byoc_synthetics.config.type, "api")
  subtype   = try(local.byoc_synthetics.config.subtype, "http")
  status    = try(local.byoc_synthetics.config.status, "paused")
  message   = try(local.byoc_synthetics.config.message, "BYOC ctl-api check")
  locations = try(local.byoc_synthetics.config.locations, ["aws:us-east-1"])
  tags      = try(local.byoc_synthetics.config.tags, [])

  request_definition {
    method = try(local.byoc_synthetics.config.request_definition.method, "GET")
    url    = "https://api.${each.value.root_domain}/livez"
  }

  request_headers = try(local.byoc_synthetics.config.request_headers, {})

  assertion {
    type     = try(local.byoc_synthetics.config.assertion.type, "statusCode")
    operator = try(local.byoc_synthetics.config.assertion.operator, "is")
    target   = try(local.byoc_synthetics.config.assertion.target, "200")
  }

  options_list {
    tick_every = try(local.byoc_synthetics.config.options_list.tick_every, 900)
    retry {
      count    = try(local.byoc_synthetics.config.options_list.retry.count, 2)
      interval = try(local.byoc_synthetics.config.options_list.retry.interval, 300)
    }
    monitor_options {
      renotify_interval = try(local.byoc_synthetics.config.options_list.monitor_options.renotify_interval, 120)
    }
    monitor_priority = try(local.byoc_synthetics.config.options_list.monitor_priority, 3)
  }
}

resource "datadog_synthetics_test" "byoc-dashboard-ui" {
  # Only create resources if we have valid instances
  for_each = local.instances_map

  name      = "[byoc] ${each.value.org_name}: ${each.value.instance_name}: dashboard-ui"
  type      = try(local.byoc_synthetics.config.type, "api")
  subtype   = try(local.byoc_synthetics.config.subtype, "http")
  status    = try(local.byoc_synthetics.config.status, "paused")
  message   = try(local.byoc_synthetics.config.message, "BYOC dashboard-ui check")
  locations = try(local.byoc_synthetics.config.locations, ["aws:us-east-1"])
  tags      = try(local.byoc_synthetics.config.tags, [])

  request_definition {
    method = try(local.byoc_synthetics.config.request_definition.method, "GET")
    url    = "https://app.${each.value.root_domain}/livez"
  }

  request_headers = try(local.byoc_synthetics.config.request_headers, {})

  assertion {
    type     = try(local.byoc_synthetics.config.assertion.type, "statusCode")
    operator = try(local.byoc_synthetics.config.assertion.operator, "is")
    target   = try(local.byoc_synthetics.config.assertion.target, "200")
  }

  options_list {
    tick_every = try(local.byoc_synthetics.config.options_list.tick_every, 900)
    retry {
      count    = try(local.byoc_synthetics.config.options_list.retry.count, 2)
      interval = try(local.byoc_synthetics.config.options_list.retry.interval, 300)
    }
    monitor_options {
      renotify_interval = try(local.byoc_synthetics.config.options_list.monitor_options.renotify_interval, 120)
    }
    monitor_priority = try(local.byoc_synthetics.config.options_list.monitor_priority, 3)
  }
}
