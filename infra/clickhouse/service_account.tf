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
