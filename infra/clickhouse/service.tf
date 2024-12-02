# we need a service to 1) access the clickhouse cluster internaly and 2) access it via twingate
# CHI does provide a method for creating one but we have seen it fail to re-create it a number of times
# and therefore do not have trust in it. we will allow it to exist, but we will not rely on it and
# instead use this one here for ctl-api and the ch-ui
resource "kubectl_manifest" "clickhouse_service" {
  yaml_body = yamlencode({
    "apiVersion" = "v1"
    "kind"       = "Service"
    "metadata" = {
      "annotations" = {
        "external-dns.alpha.kubernetes.io/internal-hostname" = "clickhouse.${local.zone}"
        "external-dns.alpha.kubernetes.io/ttl"               = "60"
      }
      "name"      = "clickhouse"
      "namespace" = "clickhouse"
    }
    "spec" = {
      "internalTrafficPolicy" = "Cluster"
      "ipFamilies" = [
        "IPv4",
      ]
      "ipFamilyPolicy" = "SingleStack"
      "ports" = [
        {
          "name"       = "http"
          "port"       = 8123
          "protocol"   = "TCP"
          "targetPort" = 8123
        },
        {
          "name"       = "client"
          "port"       = 9000
          "protocol"   = "TCP"
          "targetPort" = 9000
        },
      ]
      "sessionAffinity" = "None"
      "type"            = "ClusterIP"
    }
  })

  depends_on = [
    kubectl_manifest.clickhouse_installation,
  ]
}
