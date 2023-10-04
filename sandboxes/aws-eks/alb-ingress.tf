locals {
  aws_alb_controller = {
    service_account_name = "aws-load-balancer-controller"
  }
}

module "alb_controller_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "alb-controller-${local.vars.id}"

  create_role = true
  attach_load_balancer_controller_policy = true

  oidc_providers = {
    k8s = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["alb-ingress:${local.aws_alb_controller.service_account_name}"]
    }
  }
}

resource "helm_release" "alb-ingress-controller" {
  namespace        = "alb-ingress"
  create_namespace = true

  name       = "alb-ingress-controller"
  repository = "https://aws.github.io/eks-charts"
  chart      = "aws-load-balancer-controller"
  version    = "1.6.1"

  set {
    name  = "enableCertManager"
    value = "apply"
  }

  set {
    name  = "clusterName"
    value = module.eks.cluster_name
  }

  set {
    name  = "rbac.create"
    value = "true"
  }

  set {
    name  = "serviceAccount.create"
    value = "true"
  }

  set {
    name  = "serviceAccount.name"
    value = local.aws_alb_controller.service_account_name
  }

  set {
    name  = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
    value = module.alb_controller_irsa.iam_role_arn
  }

  depends_on = [
    helm_release.cert_manager,
    module.alb_controller_irsa
  ]
}
