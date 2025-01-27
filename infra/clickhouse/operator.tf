#
# Install Clickhouse Operator CRDs
#

locals {
  clickhouse_manifests = toset([
    "https://raw.githubusercontent.com/Altinity/clickhouse-operator/0.23.7/deploy/operator/clickhouse-operator-install-bundle.yaml"
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

resource "kubectl_manifest" "clickhouse_operator" {
  for_each  = local.all_manifests
  yaml_body = each.value
}
