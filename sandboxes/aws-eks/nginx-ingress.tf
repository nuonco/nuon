# NOTE(jm): creating this may lead to an SG getting deleted, which breaks deprovisioning. Trying without this.
resource "helm_release" "nginx-ingress-controller" {
  namespace        = "nginx-ingress"
  create_namespace = true

  name       = "nginx-ingress-controller"
  repository = "https://kubernetes.github.io/ingress-nginx"
  chart      = "ingress-nginx"
  version    = "4.8.0"
  timeout = 600

  set {
    name  = "rbac.create"
    value = "true"
  }

  depends_on = [
    helm_release.cert_manager,
    helm_release.alb-ingress-controller
  ]
}
