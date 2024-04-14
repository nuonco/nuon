# product project contains all workspaces for provisioning non-service parts of the product, such as the demo account,
# orgs, horizon and more.
resource "tfe_project" "dev" {
  name         = "dev"
  organization = data.tfe_organization.main.name
}

module "dev" {
  source = "./modules/workspace"

  name          = "dev"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/dev"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.dev.id

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

module "e2e-jon" {
  source = "./modules/workspace"

  name          = "e2e-jon"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/e2e-jon"
  variable_sets = ["aws-environment-credentials", "api-stage"]
  project_id    = tfe_project.dev.id

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

module "e2e-jordan" {
  source = "./modules/workspace"

  name          = "e2e-jordan"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/e2e-jordan"
  variable_sets = ["aws-environment-credentials", "api-stage"]
  project_id    = tfe_project.dev.id

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

module "e2e-nat" {
  source = "./modules/workspace"

  name          = "e2e-nat"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/e2e-nat"
  variable_sets = ["aws-environment-credentials", "api-stage"]
  project_id    = tfe_project.dev.id

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
