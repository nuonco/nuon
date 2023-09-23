# product project contains all workspaces for provisioning non-service parts of the product, such as the demo account,
# orgs, horizon and more.
resource "tfe_project" "demo" {
  name         = "demo"
  organization = data.tfe_organization.main.name
}

module "demo" {
  source = "./modules/workspace"

  name          = "demo"
  repo          = "powertoolsdev/demo"
  auto_apply    = true
  dir           = "terraform"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "demo-org-stage" {
  source = "./modules/workspace"

  name          = "demo-org-stage"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/demo-org"
  variable_sets = ["aws-environment-credentials", "api-stage"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

  // NOTE: we have to set the api token manually in the ui, so we don't leak it
  env_vars = {
    NUON_ORG_ID  = "org47liun91achn0opycy6jlke"
    NUON_API_URL = "https://ctl.stage.nuon.co"
  }

  triggered_by = [module.infra-terraform.workspace_id]
}

module "demo-org-prod" {
  source = "./modules/workspace"

  name          = "demo-org-prod"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/demo-org"
  variable_sets = ["aws-environment-credentials", "api-prod"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

  // NOTE: we have to set the api token manually in the ui, so we don't leak it
  env_vars = {
    NUON_ORG_ID = "orgtvkz1podyp9lmenx7o64usx"
  }
  triggered_by = [module.infra-terraform.workspace_id]
}

module "e2e-stage" {
  source = "./modules/workspace"

  name          = "e2e-stage"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "services/e2e/nuon"
  variable_sets = ["aws-environment-credentials", "api-stage"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

  // NOTE: we have to set the api token manually in the ui, so we don't leak it
  env_vars = {
    NUON_ORG_ID  = "org47liun91achn0opycy6jlke"
    NUON_API_URL = "https://ctl.stage.nuon.co"
  }

  vars = {
    east_1_count = 1
    east_2_count = 0
    west_2_count = 1
  }
  triggered_by = [module.infra-terraform.workspace_id]
}

module "e2e-prod" {
  source = "./modules/workspace"

  name          = "e2e-prod"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "services/e2e/nuon"
  variable_sets = ["aws-environment-credentials", "api-prod"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

  // NOTE: we have to set the api token manually in the ui, so we don't leak it
  env_vars = {
    NUON_ORG_ID = "orgtvkz1podyp9lmenx7o64usx"
  }

  vars = {
    east_1_count = 0
    east_2_count = 0
    west_2_count = 0
  }
  triggered_by = [module.infra-terraform.workspace_id]
}
