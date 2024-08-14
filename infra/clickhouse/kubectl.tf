locals {
  clickhouse_manifests = toset([
    "https://raw.githubusercontent.com/Altinity/clickhouse-operator/master/deploy/operator/clickhouse-operator-install-bundle.yaml"
  ])
}

data "http" "clickhouse_crd_raw" {
  for_each = local.clickhouse_manifests
  url = each.key
}

data "kubectl_file_documents" "clickhouse_crd_doc" {
  for_each = data.http.clickhouse_crd_raw
  content  = each.value.response_body
}

locals {
  all_manifests = merge([
    for src in data.kubectl_file_documents.clickhouse_crd_doc :
    src.manifests
  ]...)
}

provider "kubectl" {
  host                   = data.tfe_outputs.infra-eks-nuon.values.cluster_endpoint
  cluster_ca_certificate = base64decode(data.tfe_outputs.infra-eks-nuon.values.cluster_certificate_authority_data)
  apply_retry_count      = 3
  load_config_file       = false

  dynamic "exec" {
    for_each = local.k8s_exec
    content {
      api_version = exec.value.api_version
      command     = exec.value.command
      args        = exec.value.args
    }
  }
}

resource "kubectl_manifest" "clickhouse_operator" {
  for_each  = local.all_manifests
  yaml_body = each.value
}

resource "kubectl_manifest" "namespace_clickhouse" {
  # NOTE(fd): why like this? we already have auth setup on kubectl
  yaml_body = yamlencode({
    "apiVersion" = "v1"
    "kind" = "Namespace"
    "metadata" = {
      "labels" = {
        "kubernetes.io/metadata.name" = "clickhouse"
        "name" = "clickhouse"
      }
      "name" = "clickhouse"
    }
  })
}

resource "kubectl_manifest" "nodepool_clickhouse" {
  # NodePool for clickhouse. uses taints to define what can deploy to it.
  # depends on the default EC2NodeClass in the cluster (see infra/eks)
  # https://karpenter.sh/v0.37/concepts/nodepools/

  yaml_body = yamlencode({
    "apiVersion" = "karpenter.sh/v1beta1"
    "kind" = "NodePool"
    "metadata" = {
      "name" = "clickhouse-installation"
      "namespace" = "clickhouse"
      "labels" = {
        "app" = "clickhouse-installation"
        "app.kubernetes.io/managed-by" = "terraform"
      }
    }
    "spec" = {
      "disruption" = {
        "budgets" = [
          {
            "nodes" = "50%"
          },
        ]
        "consolidateAfter"    = "30s"
        "consolidationPolicy" = "WhenEmpty"
        "expireAfter"         = "50296s"
      }
      "limits" = {
        # 5 t3a.medium boxes
        "cpu"    = 10
        "memory" = 20480
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
            "kind" = "EC2NodeClass"
            "name" = "default"
          }
          "requirements" = [
            {
              "key" = "karpenter.sh/capacity-type"
              "operator" = "In"
              "values" = [
                "spot",
                "on-demand",
              ]
            },
            {
              "key" = "node.kubernetes.io/instance-type"
              "operator" = "In"
              "values" = [
                "t3a.medium",
              ]
            },
          ]
          "taints" = [
            {
              "effect" = "NoSchedule"
              "key" = "deployment"
              "value" = "clickhouse-installation"
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

resource "kubectl_manifest" "clickhouse_installation" {
  # generated with tfk8s and the source below
  # https://github.com/Altinity/clickhouse-operator/blob/master/docs/quick_start.md
  # NOTE: uses toleration to deploy to the NodePool defined above

  yaml_body = yamlencode({
    "apiVersion" = "clickhouse.altinity.com/v1"
    "kind"       = "ClickHouseInstallation"
    "metadata" = {
      "name"      = "clickhouse-installation"
      "namespace" = "clickhouse"
    }
    "spec" = {
      "configuration" = {
        "users" = {
          "teamnuon/password_sha256_hex" = "98fec3de803abecfcfc446bb52649627a304e264485f7de57d541b6c9652ec52",
          "teamnuon/networks/ip"         = ["0.0.0.0/0"]

        }
        "clusters" = [
          {
            "name" = "simple"
            "templates" = {
              "podTemplate" = "clickhouse:${local.image_tag}"
            }
            "layout" = {
              "replicasCount" = local.replicas
              "shardsCount" = local.shards
            }
          },
        ]
      }
      "defaults" = {
        "templates" = {
          "dataVolumeClaimTemplate" = "data-volume-template"
          "logVolumeClaimTemplate"  = "log-volume-template"
        }
      }
      "templates" = {
        # we define a podTemplate to ensure the attributes for node pool selection are set
        # and so we can define the image_tag dynamically
        "podTemplates" = [{
          "name" = "clickhouse:${local.image_tag}"
          "spec" = {
            "nodeSelector" = {
              "karpenter.sh/nodepool": "clickhouse-installation"
              "clickhouse-installation": "true"
            }
            "tolerations" = [{
              "key"      = "deployment"
              "operator" = "Equal"
              "value"    = "clickhouse-installation"
              "effect"   = "NoSchedule"
            }]
            "containers" = [
              {
                "name"  = "clickhouse"
                "image" = "clickhouse/clickhouse-server:${local.image_tag}"
                "volumeMounts" = [
                  {
                    "name"      = "data-volume-template"
                    "mountPath" = "/var/lib/clickhouse"
                  },
                  {
                    "name"      = "log-volume-template"
                    "mountPath" = "/var/log/clickhouse-server"
                  }

                ]
              }
            ]
          }
        }]
        "volumeClaimTemplates" = [
          {
            "name" = "data-volume-template"
            "spec" = {
              "accessModes" = [
                "ReadWriteOnce",
              ]
              "resources" = {
                "requests" = {
                  "storage" = local.data_volume_storage
                }
              }
            }
          },
          {
            # logs are sent out to datadog so they don't have to persist long
            # NOTE(fd): we should drop this if we can
            "name" = "log-volume-template"
            "spec" = {
              "accessModes" = [
                "ReadWriteOnce",
              ]
              "resources" = {
                "requests" = {
                  "storage" = "4Gi"
                }
              }
            }
          },
        ]
      }
    }
  })

  depends_on = [
    kubectl_manifest.clickhouse_operator,
    kubectl_manifest.nodepool_clickhouse
  ]
}
