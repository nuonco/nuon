# Note: node rotation has proven problematic in the past so we are
# opting to rotate on very long timeline, or earlier but w/ scheduled downtime.

resource "kubectl_manifest" "nodepool_clickhouse" {
  # NodePool for clickhouse. uses taints to define what can deploy to it.
  # depends on the default EC2NodeClass in the cluster (see infra/eks)
  # https://karpenter.sh/v1.0/concepts/nodepools/

  yaml_body = yamlencode({
    apiVersion = "karpenter.sh/v1"
    kind       = "NodePool"
    metadata = {
      name      = "clickhouse-installation"
      namespace = "clickhouse"
      labels = {
        "app"                          = "clickhouse-installation"
        "app.kubernetes.io/managed-by" = "terraform"
        "clickhouse-installation"      = "true"
      }
    }
    spec = {
      disruption = {
        consolidationPolicy = "WhenEmpty"
        consolidateAfter    = "15m"
        budgets = [
          {
            nodes = "0" # NOTE(fd): never disrupt/rotate
          }
        ]
      }
      limits = {
        # we use the prod limits by default. stage fits comfortably.
        # prod wants 2 + 1 t3a.large boxes (1 shard, 2 replicas), (keeper)
        # we want 1 cpu per pod (min 6). we double this and add 2 more cpus for overhead
        # as our use of clickhouse grows, we can scale up to larger machines
        cpu    = "14"
        memory = "1000Gi" # no upper bound on memory: cpu count and  type is enough
      }
      template = {
        metadata = {
          labels = {
            "clickhouse-installation" = "true"
          }
        }
        spec = {
          expireAfter = local.expireAfter
          nodeClassRef = {
            group = "karpenter.k8s.aws"
            kind  = "EC2NodeClass"
            name  = "clickhouse-installation"
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
              key      = "node.kubernetes.io/instance-type"
              operator = "In"
              values = [
                "t3a.large",
              ]
            },
            {
              key      = "topology.kubernetes.io/zone"
              operator = "In"
              values   = local.availability_zones
            },

          ]
          taints = [
            {
              effect = "NoSchedule"
              key    = "installation"
              value  = "clickhouse-installation"
            },
          ]
        }
      }
    }
  })

  depends_on = [
    kubectl_manifest.namespace_clickhouse
  ]
}

resource "kubectl_manifest" "nodepool_clickhouse_keeper" {
  # NodePool for clickhouse. uses taints to define what can deploy to it.
  # depends on the default EC2NodeClass in the cluster (see infra/eks)
  # https://karpenter.sh/v1.0/concepts/nodepools/

  yaml_body = yamlencode({
    apiVersion = "karpenter.sh/v1"
    kind       = "NodePool"
    metadata = {
      name      = "clickhouse-keeper"
      namespace = "clickhouse"
      labels = {
        "app"                          = "clickhouse-keeper"
        "app.kubernetes.io/managed-by" = "terraform"
        "clickhouse-keeper"            = "true"
      }
    }
    spec = {
      disruption = {
        consolidationPolicy = "WhenEmptyOrUnderutilized"
        consolidateAfter    = "15m"
        budgets = [
          {
            nodes = "0" # NOTE(fd): never disrupt/rotate
          }
        ]
      }
      limits = {
        cpu    = "8"
        memory = "32Gi" # no upper bound on memory: cpu count and instance type is enough
      }
      template = {
        metadata = {
          labels = {
            "clickhouse-keeper" = "true"
          }
        }
        spec = {
          expireAfter = local.expireAfter
          nodeClassRef = {
            group = "karpenter.k8s.aws"
            kind  = "EC2NodeClass"
            name  = "clickhouse-keeper"
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
              key      = "node.kubernetes.io/instance-type"
              operator = "In"
              values = [
                "t3a.medium",
              ]
            },
            {
              key      = "topology.kubernetes.io/zone"
              operator = "In"
              values   = local.availability_zones
            },

          ]
          taints = [
            {
              effect = "NoSchedule"
              key    = "installation"
              value  = "clickhouse-keeper"
            },
          ]
        }
      }
    }
  })

  depends_on = [
    kubectl_manifest.namespace_clickhouse
  ]
}
