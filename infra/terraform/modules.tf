// NOTE(jm): unfortunately, we couldn't get this working with terraform.
//
// https://app.terraform.io/app/launchpaddev/workspaces/infra-terraform/runs/run-8A3XeQo2NBNE93KC
#resource "tfe_registry_module" "echo-module" {
#name         = "echo"
#organization = data.tfe_organization.main.name

#vcs_repo {
#display_identifier = "powertoolsdev/terraform-provider-echo"
#identifier         = "powertoolsdev/terraform-provider-echo"
#oauth_token_id     = data.tfe_oauth_client.github.oauth_token_id
#}
#}
