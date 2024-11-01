resource "helm_release" "reflector" {
  namespace        = "reflector"
  create_namespace = true

  name       = "reflector"
  repository = "https://emberstack.github.io/helm-charts"
  chart      = "reflector"
  version    = "v7.1.210"

  set {
    name  = "image.repository"
    value = "431927561584.dkr.ecr.us-west-2.amazonaws.com/mirror/emberstack/kubernetes-reflector"
  }

  set {
    name  = "image.tag"
    value = "7.1.210"
  }

  depends_on = [
    module.eks_aws_auth
  ]
}
