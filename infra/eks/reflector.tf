resource "helm_release" "reflector" {
  namespace        = "reflector"
  create_namespace = true

  name       = "reflector"
  repository = "https://emberstack.github.io/helm-charts"
  chart      = "reflector"
  version    = "v7.1.210"

  depends_on = [
    helm_release.cert_manager
  ]
}
