resource "nuon_terraform_module_component" "ingress" {
  name   = "Ingress"
  app_id = nuon_app.main.id

  connected_repo = {
    directory = "infra/e2e-jordan/components/ingress"
    repo      = "powertoolsdev/mono"
    branch    = "main"
  }

  var {
    name  = "vpc_id"
    value = "{{.nuon.install.outputs.vpc.id}}"
  }

  var {
    name  = "subnet_ids"
    value = "{{.nuon.install.outputs.vpc.public_subnet_ids}}"
  }
}

resource "nuon_terraform_module_component" "services" {
  name   = "Services"
  app_id = nuon_app.main.id

  connected_repo = {
    directory = "infra/e2e-jordan/components/services"
    repo      = "powertoolsdev/mono"
    branch    = "main"
  }

  var {
    name  = "cluster_arn"
    value = "{{.nuon.install.outputs.ecs_cluster.arn}}"
  }

  var {
    name  = "subnet_ids"
    value = "{{.nuon.install.outputs.vpc.private_subnet_ids}}"
  }

  var {
    name  = "region"
    value = "{{.nuon.install.outputs.account.region}}"
  }

  var {
    name  = "aws_account_id"
    value = "{{.nuon.install.outputs.account.id}}"
  }

  var {
    name  = "ingress_security_group_id"
    value = "{{.nuon.components.ingress.outputs.security_group_id}}"
  }

  var {
    name  = "vpc_id"
    value = "{{.nuon.install.outputs.vpc.id}}"
  }

  var {
    name  = "target_group_arn"
    value = "{{.nuon.components.ingress.outputs.target_group_arn}}"
  }
}
