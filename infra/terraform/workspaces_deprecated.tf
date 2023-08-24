# we plan on removing the following workspaces
module "api-prod" {
  source = "./modules/workspace"

  name       = "api-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = false
  dir        = "services/api/infra"
  vars = {
    env = "prod"
  }
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
}

module "api-gateway-stage" {
  source = "./modules/workspace"

  name       = "api-gateway-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/api-gateway/infra"
  vars = {
    env = "stage"
  }
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.services.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "api-gateway-prod" {
  source = "./modules/workspace"

  name       = "api-gateway-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/api-gateway/infra"
  vars = {
    env = "prod"
  }
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
}

module "orgs-api-stage" {
  source = "./modules/workspace"

  name       = "orgs-api-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/orgs-api/infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "orgs-api-prod" {
  source = "./modules/workspace"

  name       = "orgs-api-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/orgs-api/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.services.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "chart-common" {
  source = "./modules/workspace"

  name                            = "chart-common"
  repo                            = "powertoolsdev/chart-common"
  auto_apply                      = true
  dir                             = "infra"
  project_id                      = tfe_project.infra.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  allowed_remote_state_workspaces = ["global"]
}

module "ci-images" {
  source = "./modules/workspace"

  name          = "ci-images"
  repo          = "powertoolsdev/ci-images"
  auto_apply    = false
  dir           = "infra"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}
