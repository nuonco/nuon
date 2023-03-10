data "tfe_outputs" "infra-orgs" {
  organization = local.terraform_organization
  workspace    = "infra-orgs-${var.env}"
}

# NOTE(jdt): This isn't ideal but more elegant than hardcoding in CI
data "tfe_outputs" "infra-eks-nuon" {
  organization = local.terraform_organization
  workspace    = "infra-eks-${var.env}-nuon"
}
