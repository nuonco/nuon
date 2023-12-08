# infra project contains all workspaces for provisioning our internal infrastructure and AWS environment
resource "tfe_project" "infra" {
  name         = "infra"
  organization = data.tfe_organization.main.name
}


module "infra-artifacts" {
  source = "./modules/workspace"

  name          = "infra-artifacts"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/artifacts"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "aws" {
  source = "./modules/workspace"

  name          = "aws-org"
  repo          = "powertoolsdev/mono"
  auto_apply    = false
  dir           = "infra/aws/cdktf.out/stacks/org"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "aws-accounts" {
  source = "./modules/workspace"

  name          = "aws-accounts"
  repo          = "powertoolsdev/mono"
  auto_apply    = false
  dir           = "infra/aws/cdktf.out/stacks/accounts"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "aws-sso" {
  source = "./modules/workspace"

  name          = "aws-sso"
  repo          = "powertoolsdev/mono"
  auto_apply    = false
  dir           = "infra/aws/cdktf.out/stacks/sso"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}


module "infra-datadog-orgs-prod" {
  source = "./modules/workspace"

  name                            = "infra-datadog-orgs-prod"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/datadog"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "datadog"]
  project_id                      = tfe_project.infra.id
  vars = {
    env = "orgs-prod"
  }
  trigger_workspaces = [module.infra-eks-orgs-prod-main.workspace_id]
}

module "infra-datadog-orgs-stage" {
  source = "./modules/workspace"

  name                            = "infra-datadog-orgs-stage"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/datadog"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "datadog"]
  project_id                      = tfe_project.infra.id
  vars = {
    env = "orgs-stage"
  }
  trigger_workspaces = [module.infra-eks-orgs-stage-main.workspace_id]
}

module "infra-datadog-prod" {
  source = "./modules/workspace"

  name                            = "infra-datadog-prod"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/datadog"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "datadog"]
  project_id                      = tfe_project.infra.id
  vars = {
    env = "prod"
  }
  trigger_workspaces = [module.infra-eks-prod-nuon.workspace_id]
}

module "infra-datadog-stage" {
  source = "./modules/workspace"

  name                            = "infra-datadog-stage"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/datadog"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "datadog"]
  project_id                      = tfe_project.infra.id
  vars = {
    env = "stage"
  }
  trigger_workspaces = [module.infra-eks-stage-nuon.workspace_id]
}

module "infra-eks-orgs-prod-main" {
  source = "./modules/workspace"

  name                            = "infra-eks-orgs-prod-main"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/eks"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "twingate-api-token"]
  project_id                      = tfe_project.infra.id

  vars = {
    account = "orgs-prod"
    pool    = "main"
  }
}

module "infra-eks-orgs-stage-main" {
  source = "./modules/workspace"

  name                            = "infra-eks-orgs-stage-main"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/eks"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "twingate-api-token"]
  project_id                      = tfe_project.infra.id
  vars = {
    account = "orgs-stage"
    pool    = "main"
  }
}

module "infra-eks-prod-nuon" {
  source = "./modules/workspace"

  name                            = "infra-eks-prod-nuon"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/eks"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "twingate-api-token"]
  project_id                      = tfe_project.infra.id
  vars = {
    account = "prod"
    pool    = "nuon"
  }
}

module "infra-eks-stage-nuon" {
  source = "./modules/workspace"

  name                            = "infra-eks-stage-nuon"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/eks"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "twingate-api-token"]
  project_id                      = tfe_project.infra.id
  vars = {
    account = "stage"
    pool    = "nuon"
  }
}

module "infra-github" {
  source = "./modules/workspace"

  name                            = "infra-github"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/github"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets = ["aws-environment-credentials",
    "github-admin-powertoolsdev",
  "github-admin-nuonco"]
  project_id = tfe_project.infra.id
}

module "infra-temporal-prod" {
  source = "./modules/workspace"

  name                            = "infra-temporal-prod"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/temporal"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.infra.id
  vars = {
    env = "prod"
  }
  trigger_workspaces = [module.infra-eks-prod-nuon.workspace_id]
}

module "infra-temporal-stage" {
  source = "./modules/workspace"

  name                            = "infra-temporal-stage"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/temporal"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.infra.id
  vars = {
    env = "stage"
  }
  trigger_workspaces = [module.infra-eks-stage-nuon.workspace_id]
}

module "infra-terraform" {
  source = "./modules/workspace"

  name                            = "infra-terraform"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/terraform"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  project_id                      = tfe_project.infra.id
}

module "nuon-dns" {
  source = "./modules/workspace"

  name                            = "nuon-dns"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/dns"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.infra.id
}

module "powertools" {
  source = "./modules/workspace"

  name                            = "powertools"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/powertools"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  project_id                      = tfe_project.infra.id
}
