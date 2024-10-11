resource "kubectl_manifest" "namespace_clickhouse" {
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
