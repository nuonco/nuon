locals {
  nuon_co_domain = data.aws_route53_zone.env_root

  external_dns_zone_filters = {
    0 = "--publish-internal-services",
    1 = "--zone-id-filter=${aws_route53_zone.internal_private.id}",
    2 = "--zone-id-filter=${local.nuon_co_domain.zone_id}",
  }

  external_dns = {
    namespace     = "external-dns"
    extra_args    = local.external_dns_zone_filters
    value_file    = "values/external-dns.yaml"
    override_file = "values/external-dns-${local.workspace_trimmed}.yaml"
  }
}

module "external_dns_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "external-dns-${local.workspace_trimmed}"

  attach_external_dns_policy = true
  external_dns_hosted_zone_arns = [
    aws_route53_zone.internal_private.arn,
    data.aws_route53_zone.env_root.arn,
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
    value = local.workspace_trimmed
  }

  set {
    name  = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
    value = module.external_dns_irsa.iam_role_arn
  }

  set {
    name  = "domain_filters[0]"
    value = local.dns.zone
  }

  set {
    name  = "domain_filters[1]"
    value = local.nuon_co_domain.name
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
    fileexists(local.external_dns.override_file) ? file(local.external_dns.override_file) : "",
  ]

  depends_on = [
    module.eks
  ]
}
