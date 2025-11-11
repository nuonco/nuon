data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "aws_ecr_repository" "ecr_repository" {
  provider = aws.infra-shared-prod
  name     = local.ecr_repository
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

data "tfe_outputs" "infra-nuonctl" {
  organization = local.terraform_organization
  workspace    = "infra-nuonctl"
}
