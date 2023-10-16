resource "nuon_helm_chart_component" "signoz" {
  name       = "Signoz"
  app_id = nuon_app.real["signoz"].id
  chart_name = "signoz"

  public_repo = {
    directory = "charts/signoz"
    repo      = "https://github.com/SigNoz"
    branch    = "main"
  }

  value {
    name  = "frontend.service.annotations.\"external-dns.alpha.kubernetes.io/hostname\""
    value = "nlb.{{ .nuon.install.public_domain }}"
  }
}
