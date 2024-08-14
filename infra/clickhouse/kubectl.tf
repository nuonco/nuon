locals {
  clickhouse_manifests = toset([
    "https://raw.githubusercontent.com/Altinity/clickhouse-operator/master/deploy/operator/clickhouse-operator-install-bundle-v1beta1.yaml",
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
