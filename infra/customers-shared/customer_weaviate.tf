locals {
  weaviate_svc_annotations = {
    "service.beta.kubernetes.io/aws-load-balancer-nlb-target-type"         = "ip"
    "service.beta.kubernetes.io/aws-load-balancer-scheme"                  = "internet-facing"
    "service.beta.kubernetes.io/aws-load-balancer-target-group-attributes" = "preserve_client_ip.enabled=false"
    "service.beta.kubernetes.io/aws-load-balancer-backend-protocol"        = "tcp"
    "external-dns.alpha.kubernetes.io/hostname"                            = "weaviate.{{ .nuon.install.public_domain }}"
  }
}

resource "nuon_helm_chart_component" "weaviate" {
  provider = nuon.sandbox

  name       = "Weaviate"
  app_id     = nuon_app.sandbox["weaviate"].id
  chart_name = "weaviate"

  public_repo = {
    directory = "weaviate"
    repo      = "https://github.com/weaviate/weaviate-helm"
    branch    = "main"
  }

  dynamic "value" {
    for_each = local.weaviate_svc_annotations
    iterator = ev
    content {
      name  = ev.key
      value = ev.value
    }
  }
}
