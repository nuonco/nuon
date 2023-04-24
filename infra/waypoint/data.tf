data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "utils_deep_merge_yaml" "vars" {
  input = [
    file("vars/defaults.yaml"),
    file("vars/${var.env}.yaml"),
  ]
}

data "tfe_outputs" "infra-eks-nuon" {
  organization = local.terraform_organization
  workspace    = "infra-eks-${var.env}-${local.vars.pool}"
}
