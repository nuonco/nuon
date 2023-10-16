resource "nuon_helm_chart_component" "signoz" {
  name       = "Signoz"
  app_id = nuon_app.real["signoz"].id
  chart_name = "signoz"

  public_repo = {
    directory = "charts/signoz"
    repo      = "https://github.com/SigNoz/charts"
    branch    = "main"
  }

  value {
    name  = "frontend.service.annotations.\"external-dns.alpha.kubernetes.io/hostname\""
    value = "nlb.{{ .nuon.install.public_domain }}"
  }
}

resource "nuon_install" "signoz_install" {
  app_id = nuon_app.real["signoz"].id

  name         = "signoz-demo"
  region       = "us-east-1"
  iam_role_arn = "arn:aws:iam::949309607565:role/nuon-demo-install-access"
}
