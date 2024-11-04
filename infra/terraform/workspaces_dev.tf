# product project contains all workspaces for provisioning non-service parts of the product, such as the demo account,
# orgs, horizon and more.
resource "tfe_project" "dev" {
  name         = "dev"
  organization = data.tfe_organization.main.name
}

module "dev" {
  source = "./modules/workspace"

  name          = "dev"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/dev"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.dev.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  env_vars = {
    NUON_API_URL = local.stage.api_url
  }

  vars = {
    org_id         = local.stage.org_id
    sandbox_org_id = local.stage.sandbox_org_id
  }
  trigger_workspaces = [module.infra-terraform.workspace_id]
  trigger_prefixes   = ["services/e2e/nuon"]
}
