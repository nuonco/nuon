data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "tfe_outputs" "infra-nuonctl" {
  organization = local.terraform_organization
  workspace    = "infra-nuonctl"
}
