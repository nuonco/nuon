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
    version         = "1.0.6"
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

# update the karpenter policy
resource "aws_cloudformation_stack" "karpenter_policy_update" {
  name         = "Karpenter-${local.karpenter.cluster_name}"
  template_url = "https://raw.githubusercontent.com/aws/karpenter-provider-aws/v${local.karpenter.version}/website/content/en/preview/getting-started/getting-started-with-karpenter/cloudformation.yaml"

  parameters = {
    ClusterName = local.karpenter.cluster_name
  }

  capabilities = ["CAPABILITY_IAM"]
}


# we applied some labels manually
# this is only necessary to do once to allow the helm chart to take over the management of the crd
# docs: https://karpenter.sh/docs/troubleshooting/#helm-error-when-installing-the-karpenter-crd-chart

# install the karpenter crds: latest point version
resource "helm_release" "karpenter_crd" {
  namespace        = "karpenter"
  create_namespace = true

  chart      = "karpenter-crd"
  name       = "karpenter-crd"
  repository = "oci://public.ecr.aws/karpenter"
  version    = local.karpenter.version

  wait = true

  values = [
    yamlencode({
      karpenter_namespace = "karpenter"
    }),
  ]
  depends_on = [
    aws_cloudformation_stack.karpenter_policy_update
  ]
}


resource "helm_release" "karpenter" {
  namespace        = "karpenter"
  create_namespace = true

  chart      = "karpenter"
  name       = "karpenter"
  repository = "oci://public.ecr.aws/karpenter"
  version    = local.karpenter.version
  skip_crds  = true # CRDs are installed by helm_release.karpenter_crd

  values = [
    # https://github.com/aws/karpenter-provider-aws/blob/release-v0.37.x/charts/karpenter/values.yaml
    yamlencode({
      replicas : local.vars.managed_node_group.desired_size
      logLevel : "debug"
      settings : {
        clusterEndpoint : module.eks.cluster_endpoint
        clusterName : local.karpenter.cluster_name
      }
      # https://github.com/aws/karpenter-provider-aws/blob/release-v1.0.6/charts/karpenter/values.yaml#L99
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
      # NOTE(fd): 1.0.6 does not support webhook.serviceNamespace - keep an eye out for errors
      # https://github.com/aws/karpenter-provider-aws/blob/release-v1.0.6/charts/karpenter/values.yaml#L140
      webhook : {
        enabled : "true"
        port : 8443
      }
      serviceAccount : {
        annotations : {
          "eks.amazonaws.com/role-arn" : module.karpenter_irsa.iam_role_arn
        }
      }
      tolerations : [
        {
          key : "CriticalAddonsOnly"
          value : "exists"
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
