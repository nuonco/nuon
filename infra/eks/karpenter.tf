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
locals {
  karpenter = {
    cluster_name    = local.workspace_trimmed
    discovery_key   = "karpenter.sh/discovery"
    discovery_value = local.workspace_trimmed
  }
}

module "karpenter_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.43"

  role_name                                  = "karpenter-controller-${local.workspace_trimmed}"
  attach_karpenter_controller_policy         = true
  enable_karpenter_instance_profile_creation = true

  karpenter_controller_cluster_id = local.karpenter.cluster_name
  karpenter_controller_node_iam_role_arns = [
    module.eks.eks_managed_node_groups["karpenter"].iam_role_arn
  ]

  oidc_providers = {
    ex = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["karpenter:karpenter"]
    }
  }
}

resource "aws_iam_instance_profile" "karpenter" {
  name = "KarpenterNodeInstanceProfile-${local.workspace_trimmed}"
  role = module.eks.eks_managed_node_groups["karpenter"].iam_role_name
}

resource "helm_release" "karpenter" {
  namespace        = "karpenter"
  create_namespace = true

  chart      = "karpenter"
  name       = "karpenter"
  repository = "oci://public.ecr.aws/karpenter"
  version    = "0.37.5"

  values = [
    # https://github.com/aws/karpenter-provider-aws/blob/release-v0.37.x/charts/karpenter/values.yaml
    yamlencode({
      replicas : local.vars.managed_node_group.desired_size
      logLevel : "debug"
      settings : {
        clusterEndpoint : module.eks.cluster_endpoint
        clusterName : local.karpenter.cluster_name
      }
      webhook : {
        enabled : false
        serviceNamespace : "karpenter"
      }
      serviceAccount : {
        annotations : {
          "eks.amazonaws.com/role-arn" : module.karpenter_irsa.iam_role_arn
        }
      }
      tolerations : [
        {
          key : "CriticalAddonsOnly"
          value : "true"
          effect : "NoSchedule"
        },
      ]
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
}
