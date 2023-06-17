# product project contains all workspaces for provisioning non-service parts of the product, such as the demo account,
# orgs, horizon and more.
resource "tfe_project" "product" {
  name         = "product"
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

module "horizon-prod" {
  source = "./modules/workspace"

  name       = "horizon-prod"
  repo       = "powertoolsdev/horizon"
  auto_apply = false
  dir        = "infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials", "hashicorp-cloud-platform"]
  project_id                      = tfe_project.product.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "infra-orgs-prod" {
  source = "./modules/workspace"

  name       = "infra-orgs-prod"
  repo       = "powertoolsdev/mono"
  dir        = "infra/orgs"
  auto_apply = false
  vars = {
    env = "prod"

    deployments_bucket_name   = "nuon-org-deployments-prod"
    installations_bucket_name = "nuon-org-installations-prod"
    orgs_bucket_name          = "nuon-orgs-prod"
    secrets_bucket_name       = "nuon-org-secrets-prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  allowed_remote_state_workspaces = [
    module.orgs-api-prod.workspace_id,
    module.workers-apps-prod.workspace_id,
    module.workers-installs-prod.workspace_id,
    module.workers-instances-prod.workspace_id,
    module.workers-executors-prod.workspace_id,
    module.workers-deployments-prod.workspace_id,
    module.workers-orgs-prod.workspace_id,
    module.api-prod.workspace_id,
  ]
}

module "infra-orgs-stage" {
  source = "./modules/workspace"

  name       = "infra-orgs-stage"
  repo       = "powertoolsdev/mono"
  dir        = "infra/orgs"
  auto_apply = true
  vars = {
    env = "stage"

    deployments_bucket_name   = "nuon-org-deployments-stage"
    installations_bucket_name = "nuon-org-installations-stage"
    orgs_bucket_name          = "nuon-orgs-stage"
    secrets_bucket_name       = "nuon-org-secrets-stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  allowed_remote_state_workspaces = [
    module.orgs-api-stage.workspace_id,
    module.workers-apps-stage.workspace_id,
    module.workers-installs-stage.workspace_id,
    module.workers-instances-stage.workspace_id,
    module.workers-executors-stage.workspace_id,
    module.workers-deployments-stage.workspace_id,
    module.workers-orgs-stage.workspace_id,
    module.api-stage.workspace_id,
  ]
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
  allowed_remote_state_workspaces = [
    module.infra-orgs-prod.workspace_id,
    module.infra-orgs-stage.workspace_id,
  ]
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
}
