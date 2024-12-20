# uses locals from karpenter.tf

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
              ]
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

# https://karpenter.sh/v1.0/concepts/nodepools/
# this nodepool is created on all clusters but is only required by prod and stage
resource "kubectl_manifest" "karpenter_nodepool_public" {
  yaml_body = yamlencode({
    apiVersion = "karpenter.sh/v1" # we are on v1 now
    kind       = "NodePool"
    metadata = {
      name = "public"
    }
    spec = {
      limits = {
        cpu    = 1000
        memory = "1000Gi"
      }
      template = {
        spec = {
          # expires 7 days after creation
          expireAfter = "168h"
          nodeClassRef = {
            group = "karpenter.k8s.aws"
            kind  = "EC2NodeClass"
            name  = "public"
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
              "values"   = "t3a.large" # notice: this is hardcoded for this nodepool
            },
            {
              key      = "topology.kubernetes.io/zone"
              operator = "In"
              values = [
                "us-west-2a",
                "us-west-2b",
                "us-west-2c",
              ]
            },
          ]
        }
      }
      # https://karpenter.sh/v1.0/concepts/disruption/
      disruption = {
        consolidationPolicy = "WhenEmptyOrUnderutilized"
        consolidateAfter    = "120s"
        budgets = [
          {
            # NOTE(fd): 1. do not rotate during the work week.
            nodes    = "0",
            schedule = "0 0 * * 1,2,3,4,5" # https://crontab.guru/#0_0_*_*_1,2,3,4,5
            duration = "24h"
            reasons = [
              "Drifted",
              "Underutilized"
            ]
          },
          {
            // NOTE(fd): 2. rotate during a four hour window during the weekend.
            nodes    = "1",
            schedule = "0 9 * * 6, 7" # https://crontab.guru/#0_9_*_*_6,7
            duration = "20h"
            reasons = [
              "Empty",
              "Drifted",
              "Underutilized"
            ]
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
