locals {
  aws_alb_contoller = {
    service_account_name : "aws-load-balancer-controller"
  }
}

module "alb_controller_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "alb-controller-${local.workspace_trimmed}"

  attach_load_balancer_controller_policy = true

  oidc_providers = {
    ex = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:${local.aws_alb_contoller.service_account_name}"]
    }
  }
}



# If you were to run this by hand
# helm install aws-load-balancer-controller eks/aws-load-balancer-controller -n kube-system --set clusterName=<cluster-name> --set serviceAccount.create=false --set serviceAccount.name=aws-load-balancer-controller
resource "helm_release" "alb-ingress-controller" {
  namespace        = "kube-system"
  create_namespace = true

  name       = "aws-load-balancer-controller"
  repository = "https://aws.github.io/eks-charts"
  chart      = "aws-load-balancer-controller"
  version    = "1.6.2"

  # https://github.com/kubernetes-sigs/aws-load-balancer-controller/blob/main/helm/aws-load-balancer-controller/README.md#configuration
  set {
    name  = "enableCertManager"
    value = "apply"
  }

  set {
    name  = "clusterName"
    value = module.eks.cluster_name
  }

  set {
    name  = "region"
    value = local.vars.region
  }

  set {
    name  = "vpcId"
    value = module.vpc.vpc_id
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
    value = local.aws_alb_contoller.service_account_name
  }

  set {
    name  = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
    value = module.alb_controller_irsa.iam_role_arn
  }

  set {
    name  = "enableServiceMutatorWebhook"
    value = "false"
  }

  values = [
    yamlencode(local.vars.alb_settings)
  ]

  depends_on = [
    module.eks_aws_auth,
    helm_release.cert_manager
  ]
}
