resource "tfe_registry_module" "echo-module" {
  name         = "echo"
  organization = data.tfe_organization.main.name

vcs_repo {
    display_identifier = "powertoolsdev/terraform-provider-echo"
    identifier         = "powertoolsdev/terraform-provider-echo"
    oauth_token_id     = data.tfe_oauth_client.github.oauth_token_id
  }
}
