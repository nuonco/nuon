locals {
  temporal = {
    value_file    = "values/temporal.yaml"
    override_file = "values/${local.name}.yaml"
    namespace     = "temporal"
    frontend_url  = "temporal-frontend.${local.zone}"
    web_url       = "temporal-ui.${local.zone}"
    image_tag     = "1.20.0"
  }
}

resource "helm_release" "temporal" {
  namespace        = local.temporal.namespace
  create_namespace = true

  # TODO(jdt): dont hardcode repo
  name       = "temporal"
  chart      = "infra-temporal"
  version    = "0.20.0"
  repository = local.helm_ecr_registry_url

  values = [
    file(local.temporal.value_file),
    fileexists(local.temporal.override_file) ? file(local.temporal.override_file) : "",
    yamlencode({
      admintools = {
        image = {
          tag = local.temporal.image_tag
        }
      }

      server = {
        config = {
          persistence = {
            default = {
              sql = {
                host     = module.primary.db_instance_address
                port     = module.primary.db_instance_port
                user     = module.primary.db_instance_username
                password = module.primary.db_instance_password
              }
            }
            visibility = {
              sql = {
                host     = module.primary.db_instance_address
                port     = module.primary.db_instance_port
                user     = module.primary.db_instance_username
                password = module.primary.db_instance_password
              }
            }
          }
        }

        image = {
          tag = local.temporal.image_tag
        }

        frontend = {
          service = {
            annotations = {
              "external-dns.alpha.kubernetes.io/internal-hostname" = local.temporal.frontend_url
              "external-dns.alpha.kubernetes.io/ttl"               = "60"
            }
          }
        }
      }

      web = {
        /* image = { */
        /*   tag = local.temporal.image_tag */
        /* } */
        service = {
          annotations = {
            "external-dns.alpha.kubernetes.io/internal-hostname" = local.temporal.web_url
            "external-dns.alpha.kubernetes.io/ttl"               = "60"
          }
        }
      }

    })
  ]

}
