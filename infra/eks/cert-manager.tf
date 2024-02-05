locals {
  cert_manager = {
    name      = "cert-manager"
    namespace = "cert-manager"
  }
}

module "cert_manager_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "cert-manager-${local.workspace_trimmed}"

  attach_cert_manager_policy = true

  # NOTE: the UI does not show an ARN, but the ARN is really just the hosted zone ID, alongside the global resource
  # identifier for this resource:
  # eg: arn:aws:route53:::hostedzone/Z00748353R29S9L6C7JJV
  cert_manager_hosted_zone_arns = [
    # setup permissions for both our internal dns zone and public zone
    aws_route53_zone.internal_private.arn,
    # public zone
    data.aws_route53_zone.env_root.arn,
  ]

  oidc_providers = {
    ex = {
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
    kubectl_manifest.karpenter_provisioner,
  ]

  lifecycle {
    # destroying the release removes the CRDs and any custom resources
    # which would remove all certs issued by cert-manager
    prevent_destroy = false
  }
}
