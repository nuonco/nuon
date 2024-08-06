locals {
  karpenter = {
    cluster_name    = local.workspace_trimmed
    discovery_key   = "karpenter.sh/discovery"
    discovery_value = local.workspace_trimmed
  }
}

data "aws_ecrpublic_authorization_token" "token" {
  provider = aws.virginia
}

module "karpenter_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name                          = "karpenter-controller-${local.workspace_trimmed}"
  attach_karpenter_controller_policy = true

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

resource "helm_release" "karpenter_crd" {
  namespace        = "default"
  create_namespace = false

  chart               = "karpenter-crd"
  name                = "karpenter-crd"
  repository          = "oci://public.ecr.aws/karpenter"
  repository_username = data.aws_ecrpublic_authorization_token.token.user_name
  repository_password = data.aws_ecrpublic_authorization_token.token.password
  version             = "0.37.0"

  # NOTE(fd): we set these explicitly to explicitly manage the CRDs
  #           if these are not set, the helm release will fail when the next update is applied
  #           syntax: https://stackoverflow.com/a/70369034
  # docs: https://karpenter.sh/preview/troubleshooting/#helm-error-when-installing-the-karpenter-crd-chart

  # 1. add app.kubernetes.io/managed-by: Helm
  set {
    name  = "ec2nodeclasses\\.karpenter\\.k8s\\.aws.metadata.labels.app\\.kubernetes\\.io/managed-by"
    value = "Helm"
  }
  set {
    name  = "nodepools\\.karpenter\\.sh.metadata.labels.app.\\.kubernetes\\.io/managed-by"
    value = "Helm"
  }
  set {
    name  = "nodeclaims\\.karpenter\\.sh.metadata.labels.app\\.kubernetes\\.io/managed-by"
    value = "Helm"
  }
  # 2. add meta.helm.sh/release-name: karpenter-crd
  set {
    name  = "ec2nodeclasses\\.karpenter\\.k8s\\.aws.metadata.annotations.meta\\.helm\\.sh/release-name"
    value = "karpenter-crd"
  }
  set {
    name  = "nodepools\\.karpenter\\.sh.metadata.annotations.meta\\.helm\\.sh/release-name"
    value = "karpenter-crd"
  }
  set {
    name  = "nodeclaims\\.karpenter\\.sh.metadata.annotations.meta\\.helm\\.sh/release-name"
    value = "karpenter-crd"
  }
  # 3. add meta.helm.sh/release-namespace: karpenter
  set {
    name  = "ec2nodeclasses\\.karpenter\\.k8s\\.aws.metadata.annotations.meta\\.helm\\.sh/release-namespace"
    value = "default"
  }
  set {
    name  = "nodepools\\.karpenter\\.sh.metadata.annotations.meta\\.helm\\.sh/release-namespace"
    value = "default"
  }
  set {
    name  = "nodeclaims\\.karpenter\\.sh.metadata.annotations.meta\\.helm\\.sh/release-namespace"
    value = "default"
  }

  values = [
    # https://karpenter.sh/preview/upgrading/upgrade-guide/#crd-upgrades
    yamlencode({
      webhook: {
        enabled          : true
        serviceName      : "karpenter"
      }
    }),
  ]
}

resource "helm_release" "karpenter" {
  namespace        = "karpenter"
  create_namespace = true

  chart               = "karpenter"
  name                = "karpenter"
  repository          = "oci://public.ecr.aws/karpenter"
  repository_username = data.aws_ecrpublic_authorization_token.token.user_name
  repository_password = data.aws_ecrpublic_authorization_token.token.password
  version             = "0.37.0"

  values = [
    # https://github.com/aws/karpenter-provider-aws/blob/main/charts/karpenter/values.yaml
    yamlencode({
      replicas : local.vars.managed_node_group.desired_size
      logLevel: "debug"
      settings : {
        clusterEndpoint        : module.eks.cluster_endpoint
        clusterName            : local.karpenter.cluster_name
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
    }),
  ]
}

# "randomize" node TTLs so that all nodes across all clusters
# aren't going down simultaneously
resource "random_integer" "node_ttl" {
  min = 60 * 60 * 11 # 11 hours
  max = 60 * 60 * 17 # 17 hours

  seed = local.karpenter.cluster_name
  keepers = {
    cluster_version = local.cluster_version
  }
}

# NOTE: `Provisioner` is now a `NodePool`
# docs: https://karpenter.sh/v0.32/upgrading/v1beta1-migration/#provisioner---nodepool
# Workaround - https://github.com/hashicorp/terraform-provider-kubernetes/issues/1380#issuecomment-967022975
# use `tfk8s -M` to convert yaml to tf map
resource "kubectl_manifest" "karpenter_provisioner" {
  yaml_body = yamlencode({
    apiVersion = "karpenter.sh/v1beta1"
    kind       = "NodePool"
    metadata = {
      name = "default"
    }
    spec = {
      limits = {
        resources = {
          cpu = 1000
        }
      }
      template = {
        spec = {
          # https://karpenter.sh/v0.32/upgrading/v1beta1-migration/#provider
          nodeClassRef = {
            apiVersion = "karpenter.k8s.aws/v1beta1"
            kind       = "EC2NodeClass"
            name       = "default"
            # https://karpenter.sh/v0.32/upgrading/v1beta1-migration/#subnetselector-securitygroupselector-and-amiselector
            # the securityGroupSelector, subnetSelector, tags have moved to EC2NodeClass
          }
          requirements = [
            {
              key      = "karpenter.sh/capacity-type"
              operator = "In"
              values   = [
                "spot",
                "on-demand",
              ]
            },
            {
              "key" = "node.kubernetes.io/instance-type"
              "operator" = "In"
              "values" = local.vars.managed_node_group.instance_types
            },
          ]
        }
      }
      disruption = {
        # https://karpenter.sh/v0.32/upgrading/v1beta1-migration/#ttlsecondsafterempty
        consolidationPolicy = "WhenEmpty"
        consolidateAfter    = "30s"
        # https://karpenter.sh/v0.32/upgrading/v1beta1-migration/#ttlsecondsuntilexpired
        expireAfter         = "${random_integer.node_ttl.result}s"
      }
    }
  })

  depends_on = [
    helm_release.karpenter
  ]
}

# docs: https://karpenter.sh/v0.37/getting-started/getting-started-with-karpenter/#5-create-nodepool
# Workaround - https://github.com/hashicorp/terraform-provider-kubernetes/issues/1380#issuecomment-967022975
# use `tfk8s -M` to convert yaml to tf map
resource "kubectl_manifest" "karpenter_ec2nodeclass" {
  yaml_body = yamlencode({
    apiVersion = "karpenter.k8s.aws/v1beta1"
    kind       = "EC2NodeClass"
    metadata = {
      name = "default"
    }
    spec = {
      role = aws_iam_instance_profile.karpenter.name
      subnetSelectorTerms = {
        tags = {
          "karpenter.sh/discovery" = local.karpenter.discovery_value
        }
      }
      securityGroupSelectorTerms = {
        tags = {
          "karpenter.sh/discovery" = local.karpenter.discovery_value
        }
      }
    }
  })

  depends_on = [
    helm_release.karpenter
  ]
}
