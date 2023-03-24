data "tfe_outputs" "infra-grafana" {
  organization = local.terraform_organization
  workspace    = "infra-grafana"
}

data "twingate_groups" "engineers" {
  name = "Engineers"
}
