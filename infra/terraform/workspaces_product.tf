# product project contains all workspaces for provisioning non-service parts of the product, such as the demo account,
# orgs, horizon and more.
resource "tfe_project" "product" {
  name         = "product"
  organization = data.tfe_organization.main.name
}

module "infra-orgs-prod" {
  source = "./modules/workspace"

  name       = "infra-orgs-prod"
  repo       = "powertoolsdev/mono"
  dir        = "infra/orgs"
  auto_apply = false
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-prod-main.workspace_id]
}

module "infra-orgs-stage" {
  source = "./modules/workspace"

  name       = "infra-orgs-stage"
  repo       = "powertoolsdev/mono"
  dir        = "infra/orgs"
  auto_apply = true
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  trigger_workspaces              = [module.infra-eks-orgs-stage-main.workspace_id]
}

module "waypoint" {
  source = "./modules/workspace"

  name                            = "waypoint"
  repo                            = "powertoolsdev/waypoint"
  auto_apply                      = true
  dir                             = "infra"
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "sandboxes" {
  source = "./modules/workspace"

  name                            = "sandboxes"
  repo                            = "powertoolsdev/mono"
  auto_apply                      = true
  dir                             = "infra/sandboxes"
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
}

module "infra-waypoint-orgs-prod" {
  source = "./modules/workspace"

  name                            = "infra-waypoint-orgs-prod"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/waypoint"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  vars = {
    env = "orgs-prod"
  }
  trigger_workspaces = [module.infra-eks-orgs-prod-main.workspace_id]
}

module "infra-waypoint-orgs-stage" {
  source = "./modules/workspace"

  name                            = "infra-waypoint-orgs-stage"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/waypoint"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  vars = {
    env = "orgs-stage"
  }
  trigger_workspaces = [module.infra-eks-orgs-stage-main.workspace_id]
}
