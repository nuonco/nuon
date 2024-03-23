locals {
  temporal = {
    version = "0.33.0"
    image_tag = "1.22.6"
    value_file    = "values/temporal.yaml"
    override_file = "values/${local.name}.yaml"
    namespace     = "temporal"
    frontend_url  = "temporal-frontend.${local.zone}"
    web_url       = "temporal-ui.${local.zone}"
  }
}

resource "helm_release" "temporal" {
  namespace        = local.temporal.namespace
  create_namespace = true

  name       = "temporal"
  version    = local.temporal.version
  chart = "https://github.com/temporalio/helm-charts/releases/download/temporal-${local.temporal.version}/temporal-${local.temporal.version}.tgz"

  values = [
    file(local.temporal.value_file),
    fileexists(local.temporal.override_file) ? file(local.temporal.override_file) : "",
    yamlencode(
    {
      server = {
        image = {
          tag = local.temporal.image_tag
        }
        config = {
          persistence = {
            default = {
              sql = {
                host     = module.primary.db_instance_address
                port     = module.primary.db_instance_port
                user     = module.primary.db_instance_username
                password = local.db_password}
            }
            visibility = {
              sql = {
                host     = module.primary.db_instance_address
                port     = module.primary.db_instance_port
                user     = module.primary.db_instance_username
                password = local.db_password}
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
        }
      }

      web = {
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
