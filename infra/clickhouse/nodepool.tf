# "randomize" node TTLs so that all nodes across all clusters
# aren't going down simultaneously
resource "random_integer" "node_ttl" {
  min = 60 * 60 * 11 # 11 hours
  max = 60 * 60 * 17 # 17 hours

  seed = "${var.env}-${local.image_tag}-${local.replicas}-${local.shards}"
  keepers = {
    pseudo_version = "${var.env}-${local.image_tag}-${local.replicas}-${local.shards}"
  }
}


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
            "nodes" = "10%"
          },
          {
            "nodes" = "2"
          },
        ]
        "consolidateAfter"    = "30s"
        "consolidationPolicy" = "WhenEmpty"
        "expireAfter"         = "${random_integer.node_ttl.result}s"
      }
      "limits" = {
        # 12 + 1 t3a.large boxes (these numbers accomodate prod)
        # 2    * 13
        # 4096 * 13
        "cpu"    = 26
        "memory" = "53248Mi"
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
