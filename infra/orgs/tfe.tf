data "tfe_outputs" "infra-eks-orgs" {
  organization = local.terraform_organization
  workspace    = "infra-eks-orgs-${var.env}-main"
}

data "tfe_outputs" "sandboxes" {
  organization = local.terraform_organization
  workspace    = "sandboxes"
}
