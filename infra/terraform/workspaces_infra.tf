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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

}

module "infra-nuonctl" {
  source = "./modules/workspace"

  name          = "infra-nuonctl"
  repo          = "powertoolsdev/mono"
  auto_apply    = true
  dir           = "infra/nuonctl"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

}

module "infra-datadog-prod" {
  source = "./modules/workspace"

  name                            = "infra-datadog-prod"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/datadog"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials", "datadog"]
  project_id    = tfe_project.infra.id
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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials", "datadog"]
  project_id    = tfe_project.infra.id
  vars = {
    env = "stage"
  }
  trigger_workspaces = [module.infra-eks-stage-nuon.workspace_id]
}



module "infra-datadog-infra-shared-ci" {
  source = "./modules/workspace"

  name                            = "infra-datadog-infra-shared-ci"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/datadog"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials", "datadog"]
  project_id    = tfe_project.infra.id
  vars = {
    env = "infra-shared-ci"
  }
  trigger_workspaces = [module.infra-eks-infra-shared-ci-nuon.workspace_id]
}




module "infra-clickhouse-prod" {
  source = "./modules/workspace"

  name                            = "infra-clickhouse-prod"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/clickhouse"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id
  vars = {
    env = "prod"
  }
  trigger_workspaces = [module.infra-eks-prod-nuon.workspace_id]
}

module "infra-clickhouse-stage" {
  source = "./modules/workspace"

  name                            = "infra-clickhouse-stage"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/clickhouse"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id
  vars = {
    env = "stage"
  }
  trigger_workspaces = [module.infra-eks-stage-nuon.workspace_id]
}

module "infra-eks-runners-prod-main" {
  source = "./modules/workspace"

  name                            = "infra-eks-runners-prod-main"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/eks"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials", "twingate-api-token"]
  project_id    = tfe_project.infra.id

  vars = {
    account = "runners-prod"
    pool    = "main"
  }
}

module "infra-eks-runners-stage-main" {
  source = "./modules/workspace"

  name                            = "infra-eks-runners-stage-main"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/eks"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials", "twingate-api-token"]
  project_id    = tfe_project.infra.id
  vars = {
    account = "runners-stage"
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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials", "twingate-api-token"]
  project_id    = tfe_project.infra.id
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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials", "twingate-api-token"]
  project_id    = tfe_project.infra.id
  vars = {
    account = "stage"
    pool    = "nuon"
  }
}



module "infra-eks-infra-shared-ci-nuon" {
  source = "./modules/workspace"

  name                            = "infra-eks-infra-shared-ci-nuon"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/eks"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials", "twingate-api-token"]
  project_id    = tfe_project.infra.id
  vars = {
    account = "infra-shared-ci"
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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials",
    "github-admin-powertoolsdev",
    "github-admin-nuonco",
    "github-admin-nuonco-shared",
  ]
  project_id = tfe_project.infra.id
}

module "infra-temporal-prod" {
  source = "./modules/workspace"

  name                            = "infra-temporal-prod"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/temporal"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id
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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id
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
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  project_id = tfe_project.infra.id
}

module "nuon-dns" {
  source = "./modules/workspace"

  name                            = "nuon-dns"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/dns"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id
}

module "infra-vercel" {
  source = "./modules/workspace"

  name                            = "infra-vercel"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/vercel"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = []
  project_id    = tfe_project.infra.id
}


module "infra-on-call" {
  source = "./modules/workspace"

  name       = "infra-on-call"
  repo       = "powertoolsdev/mono"
  dir        = "infra/on-call"
  auto_apply = true

  variable_sets = []
  project_id    = tfe_project.infra.id
}


module "infra-vantage" {
  source = "./modules/workspace"

  name                            = "infra-vantage"
  repo                            = "powertoolsdev/mono"
  dir                             = "infra/vantage"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id

  variable_sets = []
  project_id    = tfe_project.infra.id
}



module "buildkit-infra-shared-ci" {
  source = "./modules/workspace"

  name       = "buildkit-infra-shared-ci"
  repo       = "powertoolsdev/mono"
  dir        = "infra/buildkit"
  auto_apply = true
  vars = {
    env = "infra-shared-ci"
  }

  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  trigger_workspaces              = [module.infra-eks-infra-shared-ci-nuon.workspace_id]
}

module "self-hosted-runners-infra-shared-ci" {
  source = "./modules/workspace"

  name       = "self-hosted-runners-infra-shared-ci"
  repo       = "powertoolsdev/mono"
  dir        = "infra/self-hosted-runners"
  auto_apply = true
  vars = {
    env = "infra-shared-ci"
  }

  variable_sets                   = ["aws-environment-credentials"]
  project_id                      = tfe_project.product.id
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  pagerduty_service_account_id    = data.tfe_organization_membership.pagerduty.user_id
  trigger_workspaces              = [module.infra-eks-infra-shared-ci-nuon.workspace_id]
}
