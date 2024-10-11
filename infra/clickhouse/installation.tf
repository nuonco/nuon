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

#
# clickhouse installation
#

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
        "settings" = {
          "logger/level" = local.logLevel
        }
        # configure to use the zookeeper nodes
        "zookeeper" = {
          "nodes" = [
            { "host" : "clickhouse-keeper-0.clickhouse-keeper-headless.clickhouse.svc.cluster.local" },
            { "host" : "clickhouse-keeper-1.clickhouse-keeper-headless.clickhouse.svc.cluster.local" },
            { "host" : "clickhouse-keeper-2.clickhouse-keeper-headless.clickhouse.svc.cluster.local" },
          ]
        }
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
                  <type>s3_plain</type>
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
          "serviceTemplate"         = "clickhouse:${local.image_tag}"
        }
      }
      "templates" = {
        # we define a clusterServiceTemplates so we can set an internal-hostname for access via twingate
        "serviceTemplates" = [{
          "name" = "clickhouse:${local.image_tag}"
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
            "topologySpreadConstraints" = [
              # spread the pods across nodes.
              {
                "maxSkew"           = 1
                "topologyKey"       = "kubernetes.io/hostname"
                "whenUnsatisfiable" = "DoNotSchedule"
                "minDomains"        = local.hosts
                "labelSelector" = {
                  "matchLabels" = {
                    # NOTE(fd): this label is automatically applied by the CRD so we can assume it exists.
                    #           that is, however, an assumption
                    "clickhouse.altinity.com/chi" = "clickhouse-installation"
                  }
                }
              },
              # spread the pods across az:
              {
                "maxSkew"           = 1
                "topologyKey"       = "kubernetes.io/hostname"
                "whenUnsatisfiable" = "DoNotSchedule"
                "minDomains"        = length(local.availability_zones)
                "labelSelector" = {
                  "matchLabels" = {
                    # NOTE(fd): this label is automatically applied by the CRD so we can assume it exists.
                    #           that is, however, an assumption
                    "clickhouse.altinity.com/chi" = "clickhouse-installation"
                  }
                }
              }
            ]
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
                "env" = [{
                  "name"  = "CLICKHOUSE_ALWAYS_RUN_INITDB_SCRIPTS"
                  "value" = "true"
                }]
                "volumeMounts" = [
                  {
                    "name"      = "data-volume-template"
                    "mountPath" = "/var/lib/clickhouse"
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
          }
        ]
      }
    }
  })

  depends_on = [
    kubectl_manifest.clickhouse_operator,
    kubectl_manifest.nodepool_clickhouse,
    kubectl_manifest.namespace_clickhouse,
    kubectl_manifest.clickhouse_keeper_installation
  ]
}
