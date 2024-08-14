data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

locals {
  template_vars = {
    "aws_account" : var.env,
    "env" : strcontains(var.env, "stage") ? "stage" : "prod"
  }
  default_vars = templatefile("vars/defaults.yaml", local.template_vars)
  env_vars     = templatefile("vars/${var.env}.yaml", local.template_vars)
}

data "utils_deep_merge_yaml" "vars" {
  input = [
    local.default_vars,
    local.env_vars,
  ]
}

data "tfe_outputs" "infra-eks-nuon" {
  organization = local.terraform_organization
  workspace    = "infra-eks-${var.env}-nuon"
}
