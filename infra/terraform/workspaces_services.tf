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
  trigger_workspaces = [module.infra-eks-stage-nuon.workspace_id, module.infra-orgs-stage.workspace_id]
  trigger_prefixes   = ["infra/modules/service"]

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  variable_sets                   = ["aws-environment-credentials", "slack-webhooks"]
  project_id                      = tfe_project.services.id
  trigger_workspaces              = [module.infra-eks-prod-nuon.workspace_id, module.infra-orgs-prod.workspace_id]
  trigger_prefixes                = ["infra/modules/service"]
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
  trigger_workspaces = [module.infra-eks-stage-nuon.workspace_id, module.infra-orgs-stage.workspace_id]
  trigger_prefixes   = ["infra/modules/service"]

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  variable_sets                   = ["aws-environment-credentials", "slack-webhooks"]
  project_id                      = tfe_project.services.id
  trigger_workspaces              = [module.infra-eks-prod-nuon.workspace_id]
  trigger_prefixes                = ["infra/modules/service"]
}

module "wiki-stage" {
  source = "./modules/workspace"

  name       = "wiki-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/wiki/infra"
  vars = {
    env       = "stage"
    tfe_token = tfe_team_token.service-accounts-stage.token
  }
  variable_sets      = ["aws-environment-credentials", "slack-webhooks"]
  project_id         = tfe_project.services.id
  trigger_workspaces = [module.infra-eks-stage-nuon.workspace_id, module.infra-orgs-stage.workspace_id]
  trigger_prefixes   = ["infra/modules/service"]

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
}

module "wiki-prod" {
  source = "./modules/workspace"

  name       = "wiki-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = false
  dir        = "services/wiki/infra"
  vars = {
    env       = "prod"
    tfe_token = tfe_team_token.service-accounts-prod.token
  }
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  variable_sets                   = ["aws-environment-credentials", "slack-webhooks"]
  project_id                      = tfe_project.services.id
  trigger_workspaces              = [module.infra-eks-prod-nuon.workspace_id, module.infra-orgs-prod.workspace_id]
  trigger_prefixes                = ["infra/modules/service"]
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
  variable_sets = [
    "aws-environment-credentials",
    "slack-webhooks",
    "api-stage",
    "canary-azure-stage"
  ]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  trigger_workspaces              = [module.infra-eks-stage-nuon.workspace_id, module.infra-orgs-stage.workspace_id]
  trigger_prefixes                = ["infra/modules/service"]
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
  variable_sets = [
    "aws-environment-credentials",
    "slack-webhooks",
    "api-prod",
    "canary-azure-prod",
  ]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  trigger_workspaces              = [module.infra-eks-prod-nuon.workspace_id, module.infra-orgs-prod.workspace_id]
  trigger_prefixes                = ["infra/modules/service"]
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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  trigger_workspaces              = [module.infra-eks-prod-nuon.workspace_id, module.infra-orgs-prod.workspace_id]
  trigger_prefixes                = ["infra/modules/service"]
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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  trigger_workspaces              = [module.infra-eks-stage-nuon.workspace_id, module.infra-orgs-stage.workspace_id]
  trigger_prefixes                = ["infra/modules/service"]
}


module "workers-infra-tests-stage" {
  source = "./modules/workspace"

  name       = "workers-infra-tests-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-infra-tests/infra"
  vars = {
    env               = "stage"
    github_install_id = "41323514"
  }
  variable_sets = [
    "aws-environment-credentials",
    "slack-webhooks",
    "api-stage",
  ]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  trigger_workspaces              = [module.infra-eks-stage-nuon.workspace_id, module.infra-orgs-stage.workspace_id]
  trigger_prefixes                = ["infra/modules/service"]
}

module "workers-infra-tests-prod" {
  source = "./modules/workspace"

  name       = "workers-infra-tests-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-infra-tests/infra"
  vars = {
    env               = "prod"
    github_install_id = "41959553"
  }
  variable_sets = [
    "aws-environment-credentials",
    "slack-webhooks",
    "api-prod",
  ]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  trigger_workspaces              = [module.infra-eks-prod-nuon.workspace_id, module.infra-orgs-prod.workspace_id]
  trigger_prefixes                = ["infra/modules/service"]
}
