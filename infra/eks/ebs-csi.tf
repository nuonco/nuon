locals {
  ebs_csi = {
    name      = "ebs-csi-controller"
    namespace = "ebs-csi-controller"
  }
}

module "ebs_csi_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name             = "ebs-csi-${local.workspace_trimmed}"
  attach_ebs_csi_policy = true

  oidc_providers = {
    k8s = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["${local.ebs_csi.name}:${local.ebs_csi.name}-sa"]
    }
  }
}

resource "helm_release" "ebs_csi" {
  namespace        = local.ebs_csi.namespace
  create_namespace = true

  name       = local.ebs_csi.name
  repository = "https://kubernetes-sigs.github.io/aws-ebs-csi-driver"
  chart      = "aws-ebs-csi-driver"
  version    = "2.16.0"

  # https://github.com/kubernetes-sigs/aws-ebs-csi-driver/blob/master/charts/aws-ebs-csi-driver/values.yaml
  values = [
    yamlencode({
      node : {
        tolerateAllTaints : true
      }
      controller : {
        k8sTagClusterId : module.eks.cluster_name
        serviceAccount : {
          annotations : {
            "eks.amazonaws.com/role-arn" : module.ebs_csi_irsa.iam_role_arn
          }
        }
        tolerations : [
          # allow deployment to run on the same nodes as the karpenter controller
          {
            key    = "karpenter.sh/controller"
            value  = "true"
            effect = "NoSchedule"
          },
          {
            key : "CriticalAddonsOnly"
            value : "true"
            effect : "NoSchedule"
          },
        ]
        topologySpreadConstraints = {
          topologyKey       = "kubernetes.io/hostname"
          whenUnsatisfiable = "ScheduleAnyway"
          labelSelector = {
            matchLabels = {
              "app.kubernetes.io/name" = "aws-ebs-csi-driver"
            }
          }
        }
      }
    }),
  ]
}
