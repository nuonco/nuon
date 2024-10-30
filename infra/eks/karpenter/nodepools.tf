# uses locals from karpenter.tf

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
        cpu    = 1000
        memory = "1000Gi"
      }
      template = {
        spec = {
          # https://karpenter.sh/v0.32/upgrading/v1beta1-migration/#provider
          nodeClassRef = {
            apiVersion = "karpenter.k8s.aws/v1beta1"
            kind       = "EC2NodeClass"
            name       = "default"
          }
          requirements = [
            {
              key      = "karpenter.sh/capacity-type"
              operator = "In"
              values = [
                "spot",
                "on-demand",
              ]
            },
            {
              "key"      = "node.kubernetes.io/instance-type"
              "operator" = "In"
              "values"   = var.instance_types
            },
            {
              key      = "topology.kubernetes.io/zone"
              operator = "In"
              values = [
                "us-west-2a",
                "us-west-2b",
                "us-west-2c",
                "us-west-2d",
              ]
            },
          ]
        }
      }
      disruption = {
        # https://karpenter.sh/v0.32/upgrading/v1beta1-migration/#ttlsecondsafterempty
        consolidationPolicy = "WhenEmpty"
        consolidateAfter    = "30s"
        # https://karpenter.sh/v0.32/upgrading/v1beta1-migration/#ttlsecondsuntilexpired
        expireAfter = "${random_integer.node_ttl.result}s"
        budgets = [
          {
            nodes = "1",
          },
          {
            # don't allow any nodes to be disrupted during work hours
            nodes    = "0",
            schedule = "0 10 * * 1,2,3,4,5" # https://crontab.guru/#0_10_*_*_1,2,3,4,5
            duration = "11h"
          },
        ]
      }
    }
  })

  depends_on = [
    helm_release.karpenter,
  ]
}
