locals {
  instances_map = {
    for item in flatten([
      for org_name, instances in local.vars.byoc_synthetics.instances : [
        for instance_name, root_domain in instances : {
          key           = "${org_name}-${instance_name}"
          org_name      = org_name
          instance_name = instance_name
          root_domain   = root_domain
        }
      ]
    ]) : item.key => item
  }
}

resource "datadog_synthetics_test" "byoc-ctl-api" {
  # combine the map (org: [install_name: root_domain, ...]) into a list of tuples [(org, install_name, root_domain)]
  for_each = local.instances_map

  name      = "[byoc] ctl-api: ${each.value.org_name}"
  type      = local.vars.byoc_synthetics.config.type
  subtype   = local.vars.byoc_synthetics.config.subtype
  status    = local.vars.byoc_synthetics.config.status
  message   = local.vars.byoc_synthetics.config.message
  locations = local.vars.byoc_synthetics.config.locations
  tags      = local.vars.byoc_synthetics.config.tags

  request_definition {
    method = local.vars.byoc_synthetics.config.request_definition.method
    url    = "https://api.${each.value.root_domain}/livez"
  }

  request_headers = local.vars.byoc_synthetics.config.request_headers

  assertion {
    type     = local.vars.byoc_synthetics.config.assertion.type
    operator = local.vars.byoc_synthetics.config.assertion.operator
    target   = local.vars.byoc_synthetics.config.assertion.target
  }


  options_list {
    tick_every = local.vars.byoc_synthetics.config.options_list.tick_every
    retry {
      count    = local.vars.byoc_synthetics.config.options_list.retry.count
      interval = local.vars.byoc_synthetics.config.options_list.retry.interval
    }
    monitor_options {
      renotify_interval = local.vars.byoc_synthetics.config.options_list.monitor_options.renotify_interval
    }
    monitor_priority = local.vars.byoc_synthetics.config.options_list.monitor_priority
  }
}

resource "datadog_synthetics_test" "byoc-dashboard-ui" {
  # combine the map (org: [install_name: root_domain, ...]) into a list of tuples [(org, install_name, root_domain)]
  for_each = local.instances_map

  name      = "[byoc] ctl-api: ${each.value.org_name}"
  type      = local.vars.byoc_synthetics.config.type
  subtype   = local.vars.byoc_synthetics.config.subtype
  status    = local.vars.byoc_synthetics.config.status
  message   = local.vars.byoc_synthetics.config.message
  locations = local.vars.byoc_synthetics.config.locations
  tags      = local.vars.byoc_synthetics.config.tags

  request_definition {
    method = local.vars.byoc_synthetics.config.request_definition.method
    url    = "https://app.${each.value.root_domain}/livez"
  }

  request_headers = local.vars.byoc_synthetics.config.request_headers

  assertion {
    type     = local.vars.byoc_synthetics.config.assertion.type
    operator = local.vars.byoc_synthetics.config.assertion.operator
    target   = local.vars.byoc_synthetics.config.assertion.target
  }


  options_list {
    tick_every = local.vars.byoc_synthetics.config.options_list.tick_every
    retry {
      count    = local.vars.byoc_synthetics.config.options_list.retry.count
      interval = local.vars.byoc_synthetics.config.options_list.retry.interval
    }
    monitor_options {
      renotify_interval = local.vars.byoc_synthetics.config.options_list.monitor_options.renotify_interval
    }
    monitor_priority = local.vars.byoc_synthetics.config.options_list.monitor_priority
  }
}
