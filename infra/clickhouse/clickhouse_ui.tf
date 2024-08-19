resource "kubectl_manifest" "clickhouse_ui_deployment" {
  yaml_body = yamlencode({
    "apiVersion" = "apps/v1"
    "kind" = "Deployment"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/name" = "clickhouse-ui"
      }
      "name" = "ch-ui"
      "namespace" = "clickhouse"
    }
    "spec" = {
      "replicas" = 2
      "selector" = {
        "matchLabels" = {
          "app.kubernetes.io/name" = "clickhouse-ui"
        }
      }
      "template" = {
        "metadata" = {
          "labels" = {
            "app.kubernetes.io/name" = "clickhouse-ui"
          }
        }
        "spec" = {
          "containers" = [
            {
              "image" = "ghcr.io/caioricciuti/ch-ui:latest"
              "name" = "ch-ui"
              "ports" = [
                {
                  "containerPort" = 5521
                },
              ]
              "env" = [{
                "name"  = "VITE_CLICKHOUSE_URL"
                "value" = "http://clickhouse-clickhouse-installation.clickhouse.svc.cluster.local:8123"
              },
              {
                "name"  = "VITE_CLICKHOUSE_PASS"
                "value" = "teamnuon"
              },
              {
                "name"  = "VITE_CLICKHOUSE_USER"
                "value" = "teamnuon"
              }]
            },
          ]
        }
      }
    }
  })
}

resource "kubectl_manifest" "clickhouse_ui_service" {
  yaml_body  = yamlencode({
    "apiVersion" = "v1"
    "kind" = "Service"
    "metadata" = {
      "annotations" = {
        "external-dns.alpha.kubernetes.io/internal-hostname" = "ch-ui.${local.zone}"
        "external-dns.alpha.kubernetes.io/ttl"               = "60"
      }
      "name" = "ch-ui"
      "namespace" = "clickhouse"
    }
    "spec" = {
      "ports" = [
        {
          "port" = 5521
          "protocol" = "TCP"
          "targetPort" = 5521
        },
      ]
      "selector" = {
        "app.kubernetes.io/name" = "clickhouse-ui"
      }
    }
  })
}
