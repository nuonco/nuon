data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "terraform_remote_state" "chart_common" {
  backend = "remote"

  config = {
    organization = local.terraform_organization
    workspaces = {
      name = "chart-common"
    }
  }
}

data "aws_ecr_repository" "api-gateway" {
  provider = aws.infra-shared-prod
  name     = local.name
}

data "aws_route53_zone" "public" {
  name = "${var.env}.${local.vars.public_root_domain}"
}

data "tfe_outputs" "infra-eks-nuon" {
  organization = local.terraform_organization
  workspace    = "infra-eks-${local.vars.cluster_name}"
}
