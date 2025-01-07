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

module "demo_workspace" {
  source = "./modules/workspace"

  name                            = "demo"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/demo"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = []
  project_id    = tfe_project.demo.id
}
