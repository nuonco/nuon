module "metabase-prod" {
  source = "./modules/workspace"

  name       = "metabase-prod"
  repo       = "powertoolsdev/mono"
  dir        = "infra/metabase"
  auto_apply = false
  vars = {
    env = "prod"
  }

  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  trigger_workspaces              = [module.infra-eks-runners-prod-main.workspace_id]
}

module "metabase-stage" {
  source = "./modules/workspace"

  name       = "metabase-stage"
  repo       = "powertoolsdev/mono"
  dir        = "infra/metabase"
  auto_apply = true
  vars = {
    env = "stage"
  }

  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  trigger_workspaces              = [module.infra-eks-runners-stage-main.workspace_id]
}
