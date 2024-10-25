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
      amiFamily       = "AL2"
      instanceProfile = aws_iam_instance_profile.karpenter.name # https://karpenter.sh/v0.32/concepts/nodeclasses/#specinstanceprofile
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
