resource "nuon_app" "my_eks_app" {
  name = "my_eks_app"
}

resource "nuon_app_sandbox" "main" {
  app_id            = nuon_app.my_eks_app.id
  terraform_version = "v1.6.3"
  public_repo = {
    repo      = "nuonco/sandboxes"
    branch    = "main"
    directory = "aws-eks"
  }
}

resource "nuon_app_runner" "main" {
  app_id      = nuon_app.my_eks_app.id
  runner_type = "aws-eks"
}

resource "nuon_docker_build_component" "docker_image" {
  app_id     = nuon_app.my_eks_app.id
  name       = "docker_image"
  dockerfile = "Dockerfile"
  public_repo = {
    repo      = "nuonco/guides"
    directory = "aws-eks-tutorial/components/docker-image"
    branch    = "main"
  }
}

resource "nuon_terraform_module_component" "certificate" {
  app_id = nuon_app.my_eks_app.id
  name   = "certificate"
  connected_repo = {
    repo      = "nuonco/guides"
    directory = "aws-eks-tutorial/components/certificate"
    branch    = "main"
  }
  var {
    name  = "domain_name"
    value = "introspect.{{.nuon.install.sandbox.outputs.public_domain.name}}"
  }
  var {
    name  = "zone_id"
    value = "{{.nuon.install.sandbox.outputs.public_domain.zone_id}}"
  }
}

resource "nuon_helm_chart_component" "helm_chart" {
  app_id     = nuon_app.my_eks_app.id
  name       = "helm_chart"
  chart_name = "introspect"
  public_repo = {
    repo      = "nuonco/guides"
    directory = "aws-eks-tutorial/components/helm-chart"
    branch    = "main"
  }
  value {
    name  = "image.repository"
    value = "{{.nuon.components.docker_image.image.repository.uri}}"
  }
  value {
    name  = "image.tag"
    value = "{{.nuon.components.docker_image.image.tag}}"
  }
  value {
    name  = "api.nlbs.public_domain_certificate"
    value = "{{.nuon.components.certificate.outputs.public_domain_certificate_arn}}"
  }
  value {
    name  = "api.nlbs.public_domain"
    value = "nlb.{{.nuon.install.sandbox.outputs.public_domain.name}}"
  }
  value {
    name  = "app_id"
    value = "{{.nuon.app.id}}"
  }
  value {
    name  = "org_id"
    value = "{{.nuon.org.id}}"
  }
  value {
    name  = "install_id"
    value = "{{.nuon.install.id}}"
  }
  dependencies = [
    nuon_docker_build_component.docker_image.id
  ]
}
