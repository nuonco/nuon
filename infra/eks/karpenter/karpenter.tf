#
# Karpenter Terraform
#
# we jumped from version 0.16.3 to 0.37.0. this was a pretty big change. notable changes include:
#
# 1. The `Provisioner` is not called a `NodePool`. Its responsibilities/configs are split up between
#    the `NodePool` CRD and the `EC2NodeClass` CRD.
# 2. The instance_types can be set with excessive granularity but we opt for the simplest strategy that
#    preserves our current yaml configs. If we wanted to offer multiple instance types, we could do it
#    with this same strategy unless we needed to break instance types down into very granular options.
#
resource "aws_iam_instance_profile" "karpenter" {
  name = "KarpenterNodeInstanceProfile-${var.cluster_name}"
  role = var.node_iam_role_arn
}

# module "karpenter_irsa" {
#   source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
#   version = "~> 5.43"

#   role_name                                  = "karpenter-controller-${local.workspace_trimmed}"
#   attach_karpenter_controller_policy         = true
#   enable_karpenter_instance_profile_creation = true

#   karpenter_controller_cluster_id = local.karpenter.cluster_name
#   karpenter_controller_node_iam_role_arns = [
#     module.eks.eks_managed_node_groups["karpenter"].iam_role_arn
#   ]

#   oidc_providers = {
#     ex = {
#       provider_arn               = module.eks.oidc_provider_arn
#       namespace_service_accounts = ["karpenter:karpenter"]
#     }
#   }
# }

module "karpenter" {
  source  = "terraform-aws-modules/eks/aws//modules/karpenter"
  version = "20.26.1"

  cluster_name = var.cluster_name
  namespace    = var.namespace

  create_node_iam_role = false
  node_iam_role_arn    = var.node_iam_role_arn

  enable_v1_permissions = true

  enable_irsa                     = true
  irsa_oidc_provider_arn          = var.oidc_provider_arn
  irsa_namespace_service_accounts = ["karpenter:karpenter"]
  iam_role_tags = merge(var.tags, {
    karpenter = true
  })

  queue_name = "karpenter"
}

resource "helm_release" "karpenter_crd" {
  namespace        = var.namespace
  create_namespace = false

  chart      = "karpenter-crd"
  name       = "karpenter-crd"
  repository = "oci://public.ecr.aws/karpenter"
  version    = var.karpenter_version

  wait = true

  values = [
    yamlencode({
      karpenter_namespace = var.namespace
      webhook = {
        enabled     = true
        serviceName = "karpenter"
        port        = 8443
      }
    }),
  ]

  depends_on = [
    module.karpenter
  ]
}

resource "helm_release" "karpenter" {
  namespace        = var.namespace
  create_namespace = false

  chart      = "karpenter"
  name       = "karpenter"
  repository = "oci://public.ecr.aws/karpenter"
  version    = var.karpenter_version

  # https://github.com/aws/karpenter-provider-aws/blob/v1.0.6/charts/karpenter/values.yaml
  values = [
    yamlencode({
      replicas : 1
      logLevel : "debug"
      settings : {
        clusterEndpoint : var.cluster_endpoint
        clusterName : var.cluster_name
        interruptionQueue : module.karpenter.queue_name
        batchMaxDuration : "15s" # a little longer than the default
      }
      dnsPolicy : "Default"
      controller : {
        resources : {
          requests : {
            cpu : 1
            memory : "1Gi"
          }
          limits : {
            cpu : 1
            memory : "1Gi"
          }
        }
      }
      webhook : {
        enabled : "true"
        port : 8443
        serviceNamespace : "karpenter"
      }
      serviceAccount : {
        annotations : {
          "eks.amazonaws.com/role-arn" : module.karpenter.service_account
        }
      }
      tolerations : [
        {
          key : "CriticalAddonsOnly"
          value : "true"
          effect : "NoSchedule"
        },
      ]
      # https://docs.datadoghq.com/integrations/karpenter/
      podAnnotations : {
        "ad.datadoghq.com/controller.checks" : <<-EOT
          {
            "karpenter": {
              "init_config": {},
              "instances": [
                {
                  "openmetrics_endpoint": "http://%%host%%:8000/metrics"
                }
              ]
            }
          }
        EOT
      }
    }),
  ]

  depends_on = [
    helm_release.karpenter_crd
  ]
}
