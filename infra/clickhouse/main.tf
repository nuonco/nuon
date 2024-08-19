locals {
  clickhouse_manifests = toset([
    "https://raw.githubusercontent.com/Altinity/clickhouse-operator/master/deploy/operator/clickhouse-operator-install-bundle.yaml"
  ])
}

data "http" "clickhouse_crd_raw" {
  for_each = local.clickhouse_manifests
  url      = each.key
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
    "kind"       = "Namespace"
    "metadata" = {
      "labels" = {
        "kubernetes.io/metadata.name" = "clickhouse"
        "name"                        = "clickhouse"
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
            "nodes" = "50%"
          },
        ]
        "consolidateAfter"    = "30s"
        "consolidationPolicy" = "WhenEmpty"
        "expireAfter"         = "${random_integer.node_ttl.result}s"
      }
      "limits" = {
        # 5 t3a.medium boxes
        "cpu"    = 10
        "memory" = "20480Mi"
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
                "t3a.medium",
              ]
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

# configmap to bootstrap ctl_api database
resource "kubectl_manifest" "clickhouse_installation_configmap_bootstrap" {
  yaml_body = yamlencode({
    "apiVersion" = "v1"
    "kind"       = "ConfigMap"
    "metadata" = {
      "name"      = "bootstrap-configmap"
      "namespace" = "clickhouse"
    }
    "data" = {
      "01_create_databases.sh" = <<-EOT
      #!/bin/bash
      set -e
      clickhouse client -n <<-EOSQL
      CREATE DATABASE IF NOT EXISTS ctl_api;
      EOSQL

      EOT
    }
  })
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
              "podTemplate"     = "clickhouse:${local.image_tag}"
              "serviceTemplate" = "clickhouse:${local.image_tag}"
            }
            "layout" = {
              "replicasCount" = local.replicas
              "shardsCount"   = local.shards
            }
          },
        ]
        # add a storage configuration config so we can write to s3. this disk will be used for backups (/backups).
        # https://clickhouse.com/docs/en/integrations/s3#managing-credentials
        # https://clickhouse.com/docs/en/integrations/s3#configure-clickhouse-to-use-the-s3-bucket-as-a-disk
        # https://clickhouse.com/docs/en/operations/backup#configuring-backuprestore-to-use-an-s3-endpoint
        "files" = {
          "config.d/disks.xml" = <<-EOT
          <clickhouse>
            <storage_configuration>
              <disks>
                <s3_disk>
                  <type>s3</type>
                  <endpoint>https://${module.bucket.s3_bucket_bucket_domain_name}/tables/</endpoint>
                  <use_environment_credentials>true</use_environment_credentials>
                  <metadata_path>/var/lib/clickhouse/disks/s3_disk/</metadata_path>
                </s3_disk>
                <s3_cache>
                  <type>cache</type>
                  <disk>s3_disk</disk>
                  <path>/var/lib/clickhouse/disks/s3_cache/</path>
                  <max_size>10Gi</max_size>
                </s3_cache>
              </disks>
              <policies>
                <s3_main>
                  <volumes>
                    <main>
                      <disk>s3_disk</disk>
                    </main>
                  </volumes>
                </s3_main>
              </policies>
            </storage_configuration>
          </clickhouse>
          EOT
          "config.d/s3.xml"    = <<-EOT
            <clickhouse>
              <s3>
                <use_environment_credentials>true</use_environment_credentials>
              </s3>
            </clickhouse>
          EOT
        }
      }
      "defaults" = {
        "templates" = {
          "dataVolumeClaimTemplate" = "data-volume-template"
          "logVolumeClaimTemplate"  = "log-volume-template"
          "serviceTemplate" = "clickhouse:${local.image_tag}"
        }
      }
      "templates" = {
        # we define a clusterServiceTemplates so we can set an internal-hostname for access via twingate
        "serviceTemplates" = [{
          "name"     = "clickhouse:${local.image_tag}"
          "metadata" = {
            "annotations" = {
              "external-dns.alpha.kubernetes.io/internal-hostname" = "clickhouse.${local.zone}"
              "external-dns.alpha.kubernetes.io/ttl"               = "60"
            }
          }
          # default type is ClusterIP
          "spec" = {
            "ports" = [
              {
                "name" = "http"
                "port" = 8123
              },
              {
                "name" = "client"
                "port" = 9000
              }
            ]
          }
        }]
        # we define a podTemplate to ensure the attributes for node pool selection are set
        # and so we can define the image_tag dynamically
        "podTemplates" = [{
          "name" = "clickhouse:${local.image_tag}"
          "spec" = {
            "nodeSelector" = {
              "clickhouse-installation" = "true"
            }
            "tolerations" = [{
              "key"      = "installation"
              "operator" = "Equal"
              "value"    = "clickhouse-installation"
              "effect"   = "NoSchedule"
            }]
            "containers" = [
              {
                "name"  = "clickhouse"
                "image" = "clickhouse/clickhouse-server:${local.image_tag}"
                "env"   = [{
                  "name"  = "CLICKHOUSE_ALWAYS_RUN_INITDB_SCRIPTS"
                  "value" = "true"
                }]
                "volumeMounts" = [
                  {
                    "name"      = "data-volume-template"
                    "mountPath" = "/var/lib/clickhouse"
                  },
                  {
                    "name"      = "log-volume-template"
                    "mountPath" = "/var/log/clickhouse-server"
                  },
                  {
                    "name"      = "bootstrap-configmap-volume"
                    "mountPath" = "/docker-entrypoint-initdb.d"
                  }
                ],
              }
            ]
            "volumes" = [
              {
                "name" = "bootstrap-configmap-volume"
                "configMap" = {
                  "name" : "bootstrap-configmap"
                }
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

# we grab this default ServiceAccount that is created automatically by the CRD
# and declare it explicitly so we can add the eks role arn annotation for the role
# assumption stuff
resource "kubectl_manifest" "clickhouse_serviceaccount_default" {
  yaml_body = yamlencode({
    "apiVersion" = "v1"
    "kind"       = "ServiceAccount"
    "metadata" = {
      "name"      = "default"
      "namespace" = "clickhouse"
      "annotations" = {
        "eks.amazonaws.com/role-arn" = aws_iam_role.clickhouse_role.arn
      }
    }
  })
  depends_on = [
    kubectl_manifest.clickhouse_installation
  ]
}
