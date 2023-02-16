data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "aws_ecr_repository" "ecr_repository" {
  provider = aws.infra-shared-prod
  name     = local.name
}

data "terraform_remote_state" "chart_common" {
  backend = "remote"

  config = {
    organization = "launchpaddev"
    workspaces = {
      name = "chart-common"
    }
  }
}

data "tfe_outputs" "infra-orgs" {
  organization = local.terraform_organization
  workspace    = "infra-orgs-${var.env}"
}


# NOTE(jdt): This isn't ideal but more elegant than hardcoding in CI
data "tfe_outputs" "infra-eks-nuon" {
  organization = local.terraform_organization
  workspace    = "infra-eks-${var.env}-nuon"
}
