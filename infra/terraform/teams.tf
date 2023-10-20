resource "tfe_team" "service-accounts-stage" {
  name         = "service-accounts-stage"
  organization = "nuonco"
}

resource "tfe_team_access" "infra-orgs-stage" {
  access       = "read"
  team_id      = tfe_team.service-accounts-stage.id
  workspace_id = module.infra-orgs-stage.workspace_id
}

resource "tfe_team_token" "service-accounts-stage" {
  team_id = tfe_team.service-accounts-stage.id
}

resource "tfe_team" "service-accounts-prod" {
  name         = "service-accounts-prod"
  organization = "nuonco"
}

resource "tfe_team_access" "infra-orgs-prod" {
  access       = "read"
  team_id      = tfe_team.service-accounts-prod.id
  workspace_id = module.infra-orgs-prod.workspace_id
}

resource "tfe_team_token" "service-accounts-prod" {
  team_id = tfe_team.service-accounts-prod.id
}
