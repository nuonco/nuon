locals {
  internal_domain = aws_route53_zone.internal.name
  public_domain = aws_route53_zone.public.name
  external_dns = {
    namespace = "external-dns"
    extra_args = {
      0 = "--publish-internal-services",
      1 = "--zone-id-filter=${aws_route53_zone.internal.id}",
      2 = "--zone-id-filter=${aws_route53_zone.public.id}",
    }
    value_file    = "values/external-dns.yaml"
  }
}


module "external_dns_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "external-dns-${var.nuon_id}"

  attach_external_dns_policy = true
  external_dns_hosted_zone_arns = [
    aws_route53_zone.internal.arn,
    aws_route53_zone.public.arn,
  ]

  oidc_providers = {
    ex = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["${local.external_dns.namespace}:external-dns"]
    }
  }
}

resource "helm_release" "external_dns" {
  namespace        = "external-dns"
  create_namespace = true

  name       = "external-dns"
  repository = "https://kubernetes-sigs.github.io/external-dns/"
  chart      = "external-dns"
  version    = "1.12.0"

  set {
    name  = "txt_owner_id"
    value = var.nuon_id
  }

  set {
    name  = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
    value = module.external_dns_irsa.iam_role_arn
  }

  set {
    name  = "domain_filters[0]"
    value = local.internal_domain
  }

  set {
    name  = "domain_filters[1]"
    value = local.public_domain
  }

  dynamic "set" {
    for_each = local.external_dns.extra_args
    content {
      name  = "extraArgs[${set.key}]"
      value = set.value
    }
  }

  values = [
    file(local.external_dns.value_file),
  ]

  depends_on = [
     module.external_dns_irsa,
     helm_release.metrics_server
  ]
}
