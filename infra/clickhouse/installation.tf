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
      CREATE DATABASE IF NOT EXISTS ctl_api ON CLUSTER 'simple';
      EOSQL

      EOT
    }
  })
}

resource "kubectl_manifest" "clickhouse_installation" {
  # generated with tfk8s and the source below
  # https://github.com/Altinity/clickhouse-operator/blob/master/docs/quick_start.md
  # NOTE: uses toleration to deploy to the NodePool defined above
  # NOTE: uses topologySpreadConstraints to distribute pods across nodes

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
              "shardsCount"   = 1
            }
          },
        ]
        "settings" = {
          "logger/level"                    = local.logLevel
          "logger/console"                  = true
          "prometheus/endpoint"             = "/metrics"
          "prometheus/port"                 = 9363
          "prometheus/metrics"              = true
          "prometheus/events"               = true
          "prometheus/asynchronous_metrics" = true
          "prometheus/status_info"          = true
          "max_concurrent_queries"          = 2500
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
          "config.d/processors_profile_log.xml" = <<-EOT
          <clickhouse>
            <processors_profile_log>
                <ttl>event_time + INTERVAL 3 DAY</ttl>
            </processors_profile_log>
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
          "imagePullPolicy" : "IfNotPresent"
          "metadata" = {
            "annotations" = {
              # https://docs.datadoghq.com/integrations/clickhouse/?tab=containerized#overview
              # https://github.com/DataDog/integrations-core/blob/master/clickhouse/datadog_checks/clickhouse/data/conf.yaml.example
              "ad.datadoghq.com/clickhouse.checks" = <<-EOT
                  {
                    "clickhouse": {
                      "init_config": {},
                      "instances": [
                        {
                          "server": "%%host%%",
                          "port": "9000",
                          "username": "teamnuon",
                          "password": "teamnuon",
                          "tags": [{"env": "${local.tags.environment}"}, {"cluster": "simple"}]
                        }
                      ]
                    }
                  }
                EOT
            }
          }
          "spec" = {
            "nodeSelector" = {
              "clickhouse-installation" = "true"
            }
            "affinity" = {
              "podAntiAffinity" = {
                "requiredDuringSchedulingIgnoredDuringExecution" = [
                  {
                    "labelSelector" = {
                      "matchLabels" = {
                        # NOTE(fd): this label is automatically applied by the CRD so we can assume it exists.
                        #           that is, however, an assumption
                        "clickhouse.altinity.com/chi" = "clickhouse-installation"
                      }
                    }
                    "topologyKey" = "kubernetes.io/hostname"
                  },
                  {
                    "labelSelector" = {
                      "matchLabels" = {
                        # NOTE(fd): this label is automatically applied by the CRD so we can assume it exists.
                        #           that is, however, an assumption
                        "clickhouse.altinity.com/chi" = "clickhouse-installation"
                      }
                    }
                    "topologyKey" = "topology.kubernetes.io/zone"
                  },
                ]
              }
            }
            "topologySpreadConstraints" = [
              # spread the pods across nodes.
              {
                "maxSkew"           = 1
                "topologyKey"       = "kubernetes.io/hostname"
                "whenUnsatisfiable" = "ScheduleAnyway"
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
                "topologyKey"       = "topology.kubernetes.io/zone"
                "whenUnsatisfiable" = "ScheduleAnyway"
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
                "image" = "431927561584.dkr.ecr.us-west-2.amazonaws.com/mirror/clickhouse/clickhouse-server:${local.image_tag}"
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
              "storageClassName" = "ebi"
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
    kubectl_manifest.clickhouse_installation_configmap_bootstrap,
    kubectl_manifest.clickhouse_keeper_installation
  ]
}
