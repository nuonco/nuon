resource "nuon_docker_build_component" "flipt" {
  name   = "image"
  app_id     = nuon_app.sandbox["flipt"].id

  dockerfile = "Dockerfile"
  public_repo = {
    directory = "."
    repo      = "flipt-io/flipt"
    branch    = "main"
  }
}

resource "nuon_helm_chart_component" "flipt" {
  provider = nuon.sandbox

  name       = "Flipt"
  app_id     = nuon_app.sandbox["flipt"].id
  chart_name = "flipt"

  public_repo = {
    directory = "charts/flipt"
    repo      = "https://github.com/flipt-io/helm-charts"
    branch    = "main"
  }

  value {
    name = "image.repository"
    value             = "{{.nuon.components.image.image.tag}}"
  }

  value {
    name = "image.tag"
    value             = "{{.nuon.components.image.image.tag}}"
  }
}

resource "nuon_install" "flipt_install" {
  provider = nuon.sandbox

  app_id = nuon_app.sandbox["flipt"].id

  name         = "flipt-demo"
  region       = "us-east-1"
  iam_role_arn = "customer-flipt"

  depends_on = [
    nuon_app_sandbox.sandbox
  ]
}

