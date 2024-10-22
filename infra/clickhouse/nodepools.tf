# Note: node rotation has proven problematic in the past so we are
# opting to rotate on very long timeline, or earlier but w/ scheduled downtime.

resource "kubectl_manifest" "nodepool_clickhouse" {
  # NodePool for clickhouse. uses taints to define what can deploy to it.
  # depends on the default EC2NodeClass in the cluster (see infra/eks)
  # https://karpenter.sh/v0.37/concepts/nodepools/

  yaml_body = yamlencode({
    "apiVersion" = "karpenter.sh/v1beta1"
    "kind"       = "NodePool"
    "metadata" = {
      "name"      = "clickhouse-installation"
      "namespace" = "clickhouse"
      "labels" = {
        "app"                          = "clickhouse-installation"
        "app.kubernetes.io/managed-by" = "terraform"
        "clickhouse-installation"      = "true"
      }
    }
    "spec" = {
      "disruption" = {
        "budgets" = [
          {
            "nodes" = "1" # only ever rotate one node at a time
          },
          {
            # never EVER rotate nodes during work hours
            "nodes"    = "0"
            "schedule" = "0 10 * * 1,2,3,4,5" # https://crontab.guru/#0_10_*_*_1,2,3,4,5
            "duration" = "11h"
          },
        ]
        "consolidateAfter"    = "30s"
        "consolidationPolicy" = "WhenEmpty"
        "expireAfter"         = "2160h"
      }
      "limits" = {
        # we use the prod limits by default. stage fits comfortably.
        # prod wants 2 + 1 t3a.large boxes (1 shard, 2 replicas), (keeper)
        # we want 1 cpu per pod (min 6). we double this and add 2 more cpus for overhead
        # as our use of clickhouse grows, we can scale up to larger machines
        "cpu"    = "14"
        "memory" = "1000Gi" # no upper bound on memory: cpu count and instance type is enough
      }
      "template" = {
        "metadata" = {
          "labels" = {
            "clickhouse-installation" = "true"
          }
        }
        "spec" = {
          "nodeClassRef" = {
            "apiVersion" = "karpenter.k8s.aws/v1beta1"
            "kind"       = "EC2NodeClass"
            "name"       = "default"
          }
          "requirements" = [
            {
              "key"      = "karpenter.sh/capacity-type"
              "operator" = "In"
              "values" = [
                "spot",
                "on-demand",
              ]
            },
            {
              "key"      = "node.kubernetes.io/instance-type"
              "operator" = "In"
              "values" = [
                "t3a.large",
                # "t3a.xlarge",
              ]
            },
            {
              "key"      = "topology.kubernetes.io/zone"
              "operator" = "In"
              "values"   = local.availability_zones
            },

          ]
          "taints" = [
            {
              "effect" = "NoSchedule"
              "key"    = "installation"
              "value"  = "clickhouse-installation"
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
  # https://karpenter.sh/v0.37/concepts/nodepools/

  yaml_body = yamlencode({
    "apiVersion" = "karpenter.sh/v1beta1"
    "kind"       = "NodePool"
    "metadata" = {
      "name"      = "clickhouse-keeper"
      "namespace" = "clickhouse"
      "labels" = {
        "app"                          = "clickhouse-keeper"
        "app.kubernetes.io/managed-by" = "terraform"
        "clickhouse-keeper"            = "true"
      }
    }
    "spec" = {
      "disruption" = {
        "budgets" = [
          {
            "nodes" = "1" # only ever rotate one node at a time
          },
          {
            # never EVER rotate nodes during work hours
            "nodes"    = "0"
            "schedule" = "0 10 * * 1,2,3,4,5" # https://crontab.guru/#0_10_*_*_1,2,3,4,5
            "duration" = "11h"
          },
        ]
        "consolidateAfter"    = "30s"
        "consolidationPolicy" = "WhenEmpty"
        "expireAfter"         = "2160h"
      }
      "limits" = {
        # we need 3 keepers (2 cpu's per box, 1 cpu per pod)
        "cpu"    = "6"
        "memory" = "100Gi" # high upper bound on memory: cpu count and instance type is enough
      }
      "template" = {
        "metadata" = {
          "labels" = {
            "clickhouse-keeper" = "true"
          }
        }
        "spec" = {
          "nodeClassRef" = {
            "apiVersion" = "karpenter.k8s.aws/v1beta1"
            "kind"       = "EC2NodeClass"
            "name"       = "default"
          }
          "requirements" = [
            {
              "key"      = "karpenter.sh/capacity-type"
              "operator" = "In"
              "values" = [
                "spot",
                "on-demand",
              ]
            },
            {
              "key"      = "node.kubernetes.io/instance-type"
              "operator" = "In"
              "values" = [
                "t3a.medium",
              ]
            },
            {
              "key"      = "topology.kubernetes.io/zone"
              "operator" = "In"
              "values"   = local.availability_zones
            },

          ]
          "taints" = [
            {
              "effect" = "NoSchedule"
              "key"    = "installation"
              "value"  = "clickhouse-keeper"
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
