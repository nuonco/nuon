locals {
  temporal = {
    version       = "0.33.0"
    image_tag     = "${local.vars.image_tag}"
    value_file    = "values/temporal.yaml"
    override_file = "values/${local.name}.yaml"
    namespace     = "temporal"
    frontend_url  = "temporal-frontend.${local.zone}"
    web_url       = "temporal-ui.${local.zone}"
  }
  environment = local.tags.environment
}

resource "helm_release" "temporal" {
  namespace        = local.temporal.namespace
  create_namespace = true

  name    = "temporal"
  version = local.temporal.version
  chart   = "https://github.com/temporalio/helm-charts/releases/download/temporal-${local.temporal.version}/temporal-${local.temporal.version}.tgz"

  values = [
    file(local.temporal.value_file),
    fileexists(local.temporal.override_file) ? file(local.temporal.override_file) : "",
    yamlencode(
      {
        server = {
          image = {
            repository = "431927561584.dkr.ecr.us-west-2.amazonaws.com/mirror/temporalio/server"
            tag        = local.temporal.image_tag
          }

          config = {
            persistence = {
              default = {
                sql = {
                  host     = module.primary.db_instance_address
                  port     = module.primary.db_instance_port
                  user     = module.primary.db_instance_username
                  password = local.db_password
                }
              }
              visibility = {
                sql = {
                  host     = module.primary.db_instance_address
                  port     = module.primary.db_instance_port
                  user     = module.primary.db_instance_username
                  password = local.db_password
                }
              }
            }
          }

          frontend = {
            service = {
              annotations = {
                "external-dns.alpha.kubernetes.io/internal-hostname" = local.temporal.frontend_url
                "external-dns.alpha.kubernetes.io/ttl"               = "60"
              }
            }
            topologySpreadConstraints = [
              {
                maxSkew           = 2
                topologyKey       = "kubernetes.io/hostname"
                whenUnsatisfiable = "DoNotSchedule"
                labelSelector = merge(
                  {
                    "app.kubernetes.io/name"       = "temporal"
                    "app.kubernetes.io/component" = "frontend"
                  },
                  local.tags.environment == "stage" ? { "pool.nuon.co" = "temporal" } : {}
                )
              }
            ]
          }

          worker = {
            topologySpreadConstraints = [
              {
                maxSkew           = 2
                topologyKey       = "kubernetes.io/hostname"
                whenUnsatisfiable = "DoNotSchedule"
                labelSelector = merge(
                  {
                    "app.kubernetes.io/name"       = "temporal"
                    "app.kubernetes.io/component" = "worker"
                  },
                  local.tags.environment == "stage" ? { "pool.nuon.co" = "temporal" } : {}
                )
              }
            ]
          }

          matching = {
            topologySpreadConstraints = [
              {
                maxSkew           = 2
                topologyKey       = "kubernetes.io/hostname"
                whenUnsatisfiable = "DoNotSchedule"
                labelSelector = merge(
                  {
                    "app.kubernetes.io/name"       = "temporal"
                    "app.kubernetes.io/component" = "matching"
                  },
                  local.tags.environment == "stage" ? { "pool.nuon.co" = "temporal" } : {}
                )
              }
            ]
          }

          history = {
            topologySpreadConstraints = [
              {
                maxSkew           = 2
                topologyKey       = "kubernetes.io/hostname"
                whenUnsatisfiable = "DoNotSchedule"
                labelSelector = merge(
                  {
                    "app.kubernetes.io/name"       = "temporal"
                    "app.kubernetes.io/component" = "history"
                  },
                  local.tags.environment == "stage" ? { "pool.nuon.co" = "temporal" } : {}
                )
              }
            ]
          }
        }

        admintools = {
          image = {
            repository = "431927561584.dkr.ecr.us-west-2.amazonaws.com/mirror/temporalio/admin-tools"
            tag        = local.temporal.image_tag
          }
          topologySpreadConstraints = [
            {
              maxSkew           = 2
              topologyKey       = "kubernetes.io/hostname"
              whenUnsatisfiable = "DoNotSchedule"
              labelSelector = merge(
                {
                  "app.kubernetes.io/name"       = "temporal"
                  "app.kubernetes.io/component" = "admintools"
                },
                local.tags.environment == "stage" ? { "pool.nuon.co" = "temporal" } : {}
              )
            }
          ]
        }

        web = {
          service = {
            annotations = {
              "external-dns.alpha.kubernetes.io/internal-hostname" = local.temporal.web_url
              "external-dns.alpha.kubernetes.io/ttl"               = "60"
            }
          }
          image = {
            repository = "431927561584.dkr.ecr.us-west-2.amazonaws.com/mirror/temporalio/ui"
            tag        = "2.16.2"
          }
          topologySpreadConstraints = [
            {
              maxSkew           = 2
              topologyKey       = "kubernetes.io/hostname"
              whenUnsatisfiable = "DoNotSchedule"
              labelSelector = merge(
                {
                  "app.kubernetes.io/name"       = "temporal"
                  "app.kubernetes.io/component" = "web"
                },
                local.tags.environment == "stage" ? { "pool.nuon.co" = "temporal" } : {}
              )
            }
          ]
        }
      }
    )
  ]
}
