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

  trigger_workspaces = [module.infra-terraform.workspace_id]
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

    // NOTE: we do not actually need to/want to create installs in prod.
    disable_installs = true
  }
  trigger_workspaces = [module.infra-terraform.workspace_id]
}
