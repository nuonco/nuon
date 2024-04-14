# project services contains all of the service infra.
resource "tfe_project" "services" {
  name         = "services"
  organization = data.tfe_organization.main.name
}

module "ctl-api-stage" {
  source = "./modules/workspace"

  name       = "ctl-api-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/ctl-api/infra"
  vars = {
    env       = "stage"
    tfe_token = tfe_team_token.service-accounts-stage.token
  }
  variable_sets      = ["aws-environment-credentials", "slack-webhooks"]
  project_id         = tfe_project.services.id
  trigger_workspaces = [module.infra-eks-orgs-stage-main.workspace_id]

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "ctl-api-prod" {
  source = "./modules/workspace"

  name       = "ctl-api-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = false
  dir        = "services/ctl-api/infra"
  vars = {
    env       = "prod"
    tfe_token = tfe_team_token.service-accounts-prod.token
  }
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "slack-webhooks"]
  project_id                      = tfe_project.services.id
  trigger_workspaces              = [module.infra-eks-orgs-prod-main.workspace_id]
}

module "dashboard-ui-stage" {
  source = "./modules/workspace"

  name       = "dashboard-ui-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/dashboard-ui/infra"
  vars = {
    env       = "stage"
    tfe_token = tfe_team_token.service-accounts-stage.token
  }
  variable_sets      = ["aws-environment-credentials", "slack-webhooks"]
  project_id         = tfe_project.services.id
  trigger_workspaces = [module.infra-eks-orgs-stage-main.workspace_id]

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "dashboard-ui-prod" {
  source = "./modules/workspace"

  name       = "dashboard-ui-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = false
  dir        = "services/dashboard-ui/infra"
  vars = {
    env       = "prod"
    tfe_token = tfe_team_token.service-accounts-prod.token
  }
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "slack-webhooks"]
  project_id                      = tfe_project.services.id
  trigger_workspaces              = [module.infra-eks-orgs-prod-main.workspace_id]
}

module "workers-canary-stage" {
  source = "./modules/workspace"

  name       = "workers-canary-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-canary/infra"
  vars = {
    env               = "stage"
    github_install_id = "41323514"
  }
  variable_sets                   = ["aws-environment-credentials", "slack-webhooks", "api-stage"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-stage-main.workspace_id]
}

module "workers-canary-prod" {
  source = "./modules/workspace"

  name       = "workers-canary-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-canary/infra"
  vars = {
    env               = "prod"
    github_install_id = "41959553"
  }
  variable_sets                   = ["aws-environment-credentials", "slack-webhooks", "api-prod"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-prod-main.workspace_id]
}

module "workers-apps-prod" {
  source = "./modules/workspace"

  name       = "workers-apps-prod"
  repo       = "powertoolsdev/mono"
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
  repo       = "powertoolsdev/mono"
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

module "workers-executors-prod" {
  source = "./modules/workspace"

  name       = "workers-executors-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-executors/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-prod-main.workspace_id]
}

module "workers-executors-stage" {
  source = "./modules/workspace"

  name       = "workers-executors-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-executors/infra"
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
  repo       = "powertoolsdev/mono"
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
  repo       = "powertoolsdev/mono"
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
  repo       = "powertoolsdev/mono"
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
  repo       = "powertoolsdev/mono"
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
