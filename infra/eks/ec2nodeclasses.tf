# uses locals from karpenter.tf

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

# docs: https://karpenter.sh/v0.37/getting-started/getting-started-with-karpenter/#5-create-nodepool
resource "kubectl_manifest" "karpenter_ec2nodeclass" {
  yaml_body = yamlencode({
    apiVersion = "karpenter.k8s.aws/v1beta1"
    kind       = "EC2NodeClass"
    metadata = {
      name = "default"
    }
    spec = {
      amiFamily = "AL2"
      role      = module.eks.eks_managed_node_groups["karpenter"].iam_role_name
      subnetSelectorTerms = [
        {
          tags = {
            "karpenter.sh/discovery" = local.karpenter.discovery_value
          }
        }
      ]
      securityGroupSelectorTerms = [
        {
          tags = {
            "karpenter.sh/discovery" = local.karpenter.discovery_value
          }
        }
      ]
      tags = {
        "karpenter.sh/discovery" = local.karpenter.discovery_value
      }
    }
  })

  depends_on = [
    helm_release.karpenter
  ]
}

# we create an ec2 node class for every nodepool to avoid issues when migrating to 1.1
# see the note about "multiple NodePools with different kubelets that are referencing the same EC2NodeClass"
# see vars/default.yaml
# docs: https://karpenter.sh/docs/upgrading/v1-migration/#upgrade-procedure
resource "kubectl_manifest" "ec2nodeclass" {
  for_each = toset(local.vars.ec2nodeclasses)
  yaml_body = yamlencode({
    apiVersion = "karpenter.k8s.aws/v1beta1"
    kind       = "EC2NodeClass"
    metadata = {
      name = each.value
    }
    spec = {
      amiFamily = "AL2"
      role      = module.eks.eks_managed_node_groups["karpenter"].iam_role_name
      subnetSelectorTerms = [
        {
          tags = {
            "karpenter.sh/discovery" = local.karpenter.discovery_value
          }
        }
      ]
      securityGroupSelectorTerms = [
        {
          tags = {
            "karpenter.sh/discovery" = local.karpenter.discovery_value
          }
        }
      ]
      tags = {
        "karpenter.sh/discovery" = local.karpenter.discovery_value
      }
    }
  })

  depends_on = [
    helm_release.karpenter,
    kubectl_manifest.karpenter_ec2nodeclass
  ]
}
