module "e2e-stage" {
  source = "./modules/workspace"

  name          = "e2e-stage"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/e2e-stage"
  variable_sets = ["aws-environment-credentials", "api-stage"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

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

module "e2e-prod" {
  source = "./modules/workspace"

  name          = "e2e-prod"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/e2e-prod"
  variable_sets = ["aws-environment-credentials", "api-prod"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

  vars = {
    org_id         = local.prod.org_id
    sandbox_org_id = local.prod.sandbox_org_id
  }
  trigger_workspaces = [module.infra-terraform.workspace_id]
  trigger_prefixes   = ["services/e2e/nuon"]
}
