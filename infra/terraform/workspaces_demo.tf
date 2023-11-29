locals {
  stage = {
    org_id         = "orgzblonf9hol7jq92vkdriio4"
    sandbox_org_id = "orgvwpbd584d7v7o9x8oxqfo6b"
    api_url        = "https://ctl.stage.nuon.co"
  }

  prod = {
    org_id         = "orgtvkz1podyp9lmenx7o64usx"
    sandbox_org_id = "org1dc0615iykaaryb1txem6iw"
  }
}

# product project contains all workspaces for provisioning non-service parts of the product, such as the demo account,
# orgs, horizon and more.
resource "tfe_project" "demo" {
  name         = "demo"
  organization = data.tfe_organization.main.name
}

module "quickstart" {
  source = "./modules/workspace"

  name          = "quickstart"
  repo          = "nuonco/quickstart-test"
  auto_apply    = true
  dir           = "services/e2e/nuon"
  variable_sets = ["aws-environment-credentials", "api-prod"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

  // NOTE: we have to set the api token manually in the ui, so we don't leak it
  env_vars = {
    NUON_ORG_ID = local.prod.org_id
  }

  vars = {
    east_1_count = 0
    east_2_count = 0
    west_2_count = 0
  }
  triggered_by = [module.infra-terraform.workspace_id]
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
    NUON_ORG_ID  = "orgzblonf9hol7jq92vkdriio4"
    NUON_API_URL = "https://ctl.stage.nuon.co"
  }

  triggered_by = [module.infra-terraform.workspace_id]
}

module "demo-org-sandbox-stage" {
  source = "./modules/workspace"

  name          = "demo-org-sandbox-stage"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/demo-org"
  variable_sets = ["aws-environment-credentials", "api-stage"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

  // NOTE: we have to set the api token manually in the ui, so we don't leak it
  env_vars = {
    NUON_ORG_ID  = local.stage.org_id
    NUON_API_URL = local.stage.api_url
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
    NUON_ORG_ID = local.prod.org_id
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
    NUON_ORG_ID  = local.stage.org_id
    NUON_API_URL = local.stage.api_url
  }

  vars = {
    east_1_count = 0
    east_2_count = 0
    west_2_count = 0
  }
  triggered_by = [module.infra-terraform.workspace_id]
}

module "e2e-sandbox-stage" {
  source = "./modules/workspace"

  name          = "e2e-sandbox-stage"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "services/e2e/nuon"
  variable_sets = ["aws-environment-credentials", "api-stage"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

  // NOTE: we have to set the api token manually in the ui, so we don't leak it
  env_vars = {
    NUON_ORG_ID  = local.stage.sandbox_org_id
    NUON_API_URL = local.stage.api_url
  }

  vars = {
    east_1_count = 5
    east_2_count = 5
    west_2_count = 5
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
    NUON_ORG_ID = local.prod.org_id
  }

  vars = {
    east_1_count = 0
    east_2_count = 0
    west_2_count = 0
  }
  triggered_by = [module.infra-terraform.workspace_id]
}

module "customers-shared-stage" {
  source = "./modules/workspace"

  name          = "customers-shared-stage"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/customers-shared"
  variable_sets = ["aws-environment-credentials", "api-stage"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

  // NOTE: we have to set the api token manually in the ui, so we don't leak it
  env_vars = {
    NUON_API_URL = local.stage.api_url
  }

  vars = {
    sandbox_org_id = local.stage.sandbox_org_id
    org_id         = local.stage.org_id
  }

  triggered_by = [module.infra-terraform.workspace_id]
}

module "customers-shared-prod" {
  source = "./modules/workspace"

  name          = "customers-shared-prod"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/customers-shared"
  variable_sets = ["aws-environment-credentials", "api-prod"]
  project_id    = tfe_project.product.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url

  vars = {
    sandbox_org_id = local.prod.sandbox_org_id
    org_id         = local.prod.org_id
  }
  triggered_by = [module.infra-terraform.workspace_id]
}
