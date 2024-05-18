data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "tfe_outputs" "infra-orgs" {
  organization = local.terraform_organization
  workspace    = "infra-orgs-${var.env}"
}
