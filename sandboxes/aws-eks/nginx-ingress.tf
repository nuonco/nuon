resource "helm_release" "nginx-ingress-controller" {
  namespace        = "kube-system"
  create_namespace = true

  name       = "nginx-ingress-controller"
  repository = "https://helm.nginx.com/stable"
  chart      = "nginx-ingress"
  version    = "0.18.0"

  set {
    name  = "rbac.create"
    value = "true"
  }

  depends_on = [
    helm_release.cert_manager,
    helm_release.alb-ingress-controller
  ]
}
