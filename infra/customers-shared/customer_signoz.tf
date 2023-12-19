resource "nuon_helm_chart_component" "signoz" {
  provider = nuon.sandbox

  name       = "Signoz"
  app_id     = nuon_app.sandbox["signoz"].id
  chart_name = "signoz"

  public_repo = {
    directory = "charts/signoz"
    repo      = "https://github.com/SigNoz/charts"
    branch    = "main"
  }

  #value {
  #name  = "frontend.service.annotations.\"external-dns.alpha.kubernetes.io/hostname\""
  #value = "nlb.{{ .nuon.install.public_domain }}"
  #}
}
