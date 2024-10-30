# uses locals from karpenter.tf

# https://karpenter.sh/v1.0/concepts/nodeclasses/
resource "kubectl_manifest" "karpenter_ec2nodeclass_default" {
  yaml_body = yamlencode({
    apiVersion = "karpenter.k8s.aws/v1"
    kind       = "EC2NodeClass"
    metadata = {
      name = "default"
    }
    spec = {
      amiFamily       = "AL2"
      instanceProfile = var.node_iam_role_name
      subnetSelectorTerms = [
        {
          tags = {
            "karpenter.sh/discovery" = var.discovery_value
          }
        }
      ]
      securityGroupSelectorTerms = [
        {
          tags = {
            "karpenter.sh/discovery" = var.discovery_value
          }
        }
      ]
      tags = {
        "karpenter.sh/discovery" = var.discovery_value
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
  for_each = toset(var.ec2nodeclasses)
  yaml_body = yamlencode({
    apiVersion = "karpenter.k8s.aws/v1"
    kind       = "EC2NodeClass"
    metadata = {
      name = each.value
    }
    spec = {
      amiFamily = "AL2"
      # we use the nodegroup from the managed node group
      instanceProfile = var.node_iam_role_name
      subnetSelectorTerms = [
        {
          tags = {
            "karpenter.sh/discovery" = var.discovery_value
          }
        }
      ]
      securityGroupSelectorTerms = [
        {
          tags = {
            "karpenter.sh/discovery" = var.discovery_value
          }
        }
      ]
      tags = {
        "karpenter.sh/discovery" = var.discovery_value
      }
    }
  })

  depends_on = [
    helm_release.karpenter,
  ]
}
