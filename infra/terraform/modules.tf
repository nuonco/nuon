locals {
  # NOTE(jm): when we originally set up the github -> terraform cloud integration, we did it by using the standard app,
  # instead of a dedicated oauth connection. This does not work for managing workspaces via terraform, so as a short term
  # solution, we've made a connection tied to @jonmorehouse here - https://www.terraform.io/cloud-docs/vcs/github
  #
  # please see https://github.com/powertoolsdev/infra-terraform/issues/1
  oauth_client_id = "oc-njndoeEPx19BePSB"
}

data "tfe_oauth_client" "github" {
  oauth_client_id = local.oauth_client_id
}

data "tfe_organization" "main" {
  name = "launchpaddev"
}

resource "tfe_registry_module" "echo-module" {
  name         = "echo"
  organization = data.tfe_organization.main.name

  vcs_repo {
    display_identifier = "powertoolsdev/mono"
    identifier         = "powertoolsdev/mono"
    oauth_token_id     = data.tfe_oauth_client.github.oauth_token_id
  }
}
