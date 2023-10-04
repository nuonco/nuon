locals {
  cert_manager = {
    name      = "cert-manager"
    namespace = "cert-manager"
  }
}

module "cert_manager_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "cert-manager-${local.vars.id}"

  attach_cert_manager_policy    = true
  cert_manager_hosted_zone_arns = [
    aws_route53_zone.internal.arn,
    aws_route53_zone.public.arn,
  ]

  oidc_providers = {
    k8s = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["${local.cert_manager.namespace}:${local.cert_manager.name}"]
    }
  }
}

resource "helm_release" "cert_manager" {
  namespace        = "cert-manager"
  create_namespace = true

  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  version    = "v1.11.0"

  set {
    name  = "installCRDs"
    value = "true"
  }

  set {
    name  = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
    value = module.cert_manager_irsa.iam_role_arn
  }

  set {
    name  = "securityContext.fsGroup"
    value = "1001"
  }

  depends_on = [
    module.cert_manager_irsa,
    helm_release.external_dns
  ]
}
