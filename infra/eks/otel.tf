locals {
  otel = {
    namespace = "otel"
    instances = {
      agent = {
        value_file    = "values/otel-agent.yaml"
        override_file = "values/otel-agent-${local.workspace_trimmed}.yaml"
        extraEnvs = {
          extraEnvs : [
            {
              name : "MY_NODE_NAME"
              valueFrom : {
                fieldRef : {
                  fieldPath : "spec.nodeName"
                }
              }
            },
            { name : "HOST_PROC", value : "/hostfs/proc" },
            { name : "HOST_SYS", value : "/hostfs/sys" },
            { name : "HOST_ETC", value : "/hostfs/etc" },
            { name : "HOST_VAR", value : "/hostfs/var" },
            { name : "HOST_RUN", value : "/hostfs/run" },
            { name : "HOST_DEV", value : "/hostfs/dev" },
          ]
        }
      }
      collector = {
        value_file    = "values/otel-collector.yaml"
        override_file = "values/otel-collector-${terraform.workspace}.yaml"
        extraEnvs = {
          extraEnvs : []
        }
      }
    }
  }
}

resource "helm_release" "otel" {
  for_each         = local.otel.instances
  namespace        = local.otel.namespace
  create_namespace = true

  name       = each.key
  repository = "https://open-telemetry.github.io/opentelemetry-helm-charts"
  chart      = "opentelemetry-collector"
  version    = "0.39.2"

  values = [
    file(each.value.value_file),
    fileexists(each.value.override_file) ? file(each.value.override_file) : "",
    yamlencode(each.value.extraEnvs),

    yamlencode({
      serviceAccount = {
        name = "otel-${each.key}"
      }
    }),

    # set attributes for otel collector
    each.key != "collector" ? "" : yamlencode({
      config = {
        exporters = {
          "prometheusremotewrite/grafana" = {
            endpoint = data.tfe_outputs.infra-grafana.values.otel.prometheus_remote_write_endpoint
          }
        },
        extensions = {
          "basicauth/prometheusremotewrite" = {
            client_auth = {
              username = data.tfe_outputs.infra-grafana.values.otel.prometheus_user_id
              password = data.tfe_outputs.infra-grafana.values.otel.api_key
            }
          },
          "basicauth/otlp" = {
            client_auth = {
              # TODO(jm): expose the instance id for our stack from infra-grafana
              username = "526406"
              password = data.tfe_outputs.infra-grafana.values.otel.api_key
            }
          }
        },
        processors = {
          "attributes/default" = {
            actions = [
              {
                action = "insert"
                key    = "env"
                value  = var.account
              },
              {
                action = "insert"
                key    = "pool"
                value  = var.pool
              },
              {
                action = "insert"
                key    = "cluster"
                value  = local.workspace_trimmed
              },
              {
                action = "insert"
                key    = "region"
                value  = local.vars.region
              }
            ]
          }
        }
      }
    })
  ]

  depends_on = [
    kubectl_manifest.karpenter_provisioner,
  ]
}
