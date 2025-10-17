# uses locals from karpenter.tf

locals {
  default_zones = [
    "us-west-2a",
    "us-west-2b",
    "us-west-2c",
  ]
}

# https://karpenter.sh/v1.0/concepts/nodepools/
resource "kubectl_manifest" "karpenter_nodepool_default" {
  yaml_body = yamlencode({
    apiVersion = "karpenter.sh/v1" # we are on v1 now
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
          expireAfter = "24h"
          nodeClassRef = {
            group = "karpenter.k8s.aws"
            kind  = "EC2NodeClass"
            name  = "default"
          }
          requirements = [
            {
              key      = "karpenter.sh/capacity-type"
              operator = "In"
              values = [
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
              values   = local.default_zones
            },
          ]
        }
      }
      # https://karpenter.sh/v1.0/concepts/disruption/
      disruption = {
        consolidationPolicy = "WhenEmptyOrUnderutilized"
        consolidateAfter    = "1m"
        budgets = [
          {
            nodes = "1",
          },
          {
            # don't allow any nodes to be disrupted during work hours
            nodes    = "1",
            schedule = "0 10 * * 1,2,3,4,5" # https://crontab.guru/#0_10_*_*_1,2,3,4,5
            duration = "11h"
          },
        ]
      }
    }
  })

  depends_on = [
    kubectl_manifest.ec2nodeclass,
    helm_release.karpenter,
  ]
}

# dynamic aditional nodepools from env file
resource "kubectl_manifest" "additional_nodepools" {
  for_each = { for np in var.additional_nodepools : np.name => np }

  yaml_body = yamlencode({
    apiVersion = "karpenter.sh/v1" # we are on v1 now
    kind       = "NodePool"
    metadata = {
      name = each.value.name
      labels = lookup(each.value, "labels", {
        "pool.nuon.co" : each.value.name
      })
    }
    spec = {
      limits = {
        cpu    = each.value.limits.cpu
        memory = each.value.limits.memory
      }
      template = {
        metadata = {
          labels = lookup(each.value, "labels", {
            "pool.nuon.co" : each.value.name
          })
        }
        spec = {
          expireAfter = each.value.expireAfter
          nodeClassRef = {
            group = "karpenter.k8s.aws"
            kind  = "EC2NodeClass"
            name  = lookup(each.value, "nodeclass", each.value.name)
          }
          requirements = [
            {
              key      = "karpenter.sh/capacity-type"
              operator = "In"
              values = [
                "on-demand",
              ]
            },
            {
              "key"      = "node.kubernetes.io/instance-type"
              "operator" = "In"
              "values"   = each.value.instance_types
            },
            {
              key      = "topology.kubernetes.io/zone"
              operator = "In"
              values   = lookup(each.value, "zones", local.default_zones)
            },
          ]
          taints = lookup(each.value, "taints", [
            {
              key    = "pool.nuon.co"
              value  = each.value.name
              effect = "NoSchedule"
            }
          ])
        }
      }
      # https://karpenter.sh/v1.0/concepts/disruption/
      disruption = each.value.disruption
    }
  })

  depends_on = [
    kubectl_manifest.ec2nodeclass,
    helm_release.karpenter,
  ]
}

