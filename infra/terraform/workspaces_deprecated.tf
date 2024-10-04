module "workers-apps-prod" {
  source = "./modules/workspace"

  name       = "workers-apps-prod"
  repo       = ""
  auto_apply = true
  dir        = "services/workers-apps/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-prod-main.workspace_id]
}

module "workers-apps-stage" {
  source = "./modules/workspace"

  name       = "workers-apps-stage"
  repo       = ""
  auto_apply = true
  dir        = "services/workers-apps/infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-stage-main.workspace_id]
}

module "workers-installs-prod" {
  source = "./modules/workspace"

  name       = "workers-installs-prod"
  repo       = ""
  auto_apply = true
  dir        = "services/workers-installs/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-prod-main.workspace_id]
}

module "workers-installs-stage" {
  source = "./modules/workspace"

  name       = "workers-installs-stage"
  repo       = ""
  auto_apply = true
  dir        = "services/workers-installs/infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-stage-main.workspace_id]
}

module "workers-orgs-prod" {
  source = "./modules/workspace"

  name       = "workers-orgs-prod"
  repo       = ""
  auto_apply = true
  dir        = "services/workers-orgs/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-prod-main.workspace_id]
}

module "workers-orgs-stage" {
  source = "./modules/workspace"

  name       = "workers-orgs-stage"
  repo       = ""
  auto_apply = true
  dir        = "services/workers-orgs/infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-stage-main.workspace_id]
}
