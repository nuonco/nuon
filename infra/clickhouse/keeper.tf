#
# clickhouse keeper installation
#
resource "kubectl_manifest" "clickhouse_keeper_installation" {
  yaml_body = yamlencode({
    "apiVersion" = "clickhouse-keeper.altinity.com/v1"
    "kind"       = "ClickHouseKeeperInstallation"
    "metadata" = {
      "name"      = "clickhouse-keeper"
      "namespace" = "clickhouse"
    }
    "spec" = {
      "configuration" = {
        "clusters" = [
          {
            "layout" = {
              "replicasCount" = local.keeperReplicaCount
            }
            "name" = "chk-simple"
          },
        ]
        "settings" = {
          "keeper_server/coordination_settings/raft_logs_level" = "information"
          "keeper_server/four_letter_word_white_list"           = "*"
          "keeper_server/raft_configuration/server/port"        = "9444"
          "keeper_server/storage_path"                          = "/var/lib/clickhouse-keeper"
          "keeper_server/tcp_port"                              = "2181"
          "listen_host"                                         = "0.0.0.0"
          "logger/console"                                      = "true"
          "logger/level"                                        = "debug"
          "prometheus/asynchronous_metrics"                     = "true"
          "prometheus/endpoint"                                 = "/metrics"
          "prometheus/events"                                   = "true"
          "prometheus/metrics"                                  = "true"
          "prometheus/port"                                     = "7000"
          "prometheus/status_info"                              = "false"
        }
      }
      "templates" = {
        "serviceTemplates" = [{
          "name" = "clickhouse-keeper:${local.image_tag}"
          "metadata" = {
            "namespace" = "clickhouse"
          }
          "spec" = {
            "ports" = [
              {
                "name"       = "tcp"
                "port"       = 2181
                "targetPort" = 2181
              }
            ]
          }
        }]
        "podTemplates" = [
          {
            "name" = "clickhouse-keeper:${local.image_tag}"
            "spec" = {
              "securityContext" = {
                "fsGroup"   = 101
                "runAsUser" = 101
              }
              "topologySpreadConstraints" = [
                # spread the pods across nodes.
                {
                  "maxSkew"           = 1
                  "topologyKey"       = "kubernetes.io/hostname"
                  "whenUnsatisfiable" = "DoNotSchedule"
                  "minDomains"        = local.keeperReplicaCount
                  "labelSelector" = {
                    "matchLabels" = {
                      # NOTE(fd): this label is automatically applied by the CRD so we can assume it exists.
                      #           that is, however, an assumption
                      "app" = "clickhouse-keeper"
                    }
                  }
                }
              ]
              "tolerations" = [{
                "key"      = "installation"
                "operator" = "Equal"
                "value"    = "clickhouse-keeper"
                "effect"   = "NoSchedule"
              }]
              "containers" = [
                {
                  "image"           = "clickhouse/clickhouse-keeper:24.3.12.75-alpine"
                  "imagePullPolicy" = "IfNotPresent"
                  "name"            = "clickhouse-keeper"
                  "resources" = {
                    "limits" = {
                      "cpu"    = "2"
                      "memory" = "4Gi"
                    }
                    "requests" = {
                      "cpu"    = "1"
                      "memory" = "256M"
                    }
                  }
                },
              ]
            }
          }
        ]
      }
    }
  })

  depends_on = [
    kubectl_manifest.clickhouse_operator,
    kubectl_manifest.namespace_clickhouse,
    kubectl_manifest.nodepool_clickhouse_keeper,
  ]
}
