# each cluster gets a wild card cert that any service in the cluster can use
resource "kubectl_manifest" "wildcard_cert" {
  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"

    metadata = {
      name      = "wildcard"
      namespace = "default"
      labels = {
        "app.kubernetes.io/name"       = "wildcard"
        "app.kubernetes.io/managed-by" = "infra-eks"
      }
    }

    spec = {
      secretName = "wildcard-tls"
      dnsNames = [
        "*.${data.aws_route53_zone.env_root.name}",
        "${data.aws_route53_zone.env_root.name}"
      ]
      issuerRef = {
        name = local.cert_manager_issuers.public_issuer_name
        kind = "ClusterIssuer"
      }
      secretTemplate = {
        annotations = {
          "reflector.v1.k8s.emberstack.com/reflection-allowed"            = "true"
          "reflector.v1.k8s.emberstack.com/reflection-allowed-namespaces" = ""
        }
      }
    }
  })

  depends_on = [
    helm_release.cert_manager,
  ]
}
