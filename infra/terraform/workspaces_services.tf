module "api-stage" {
  source = "./modules/workspace"

  name       = "api-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/api/infra"
  vars = {
    env = "stage"
  }
  variable_sets = ["aws-environment-credentials"]

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

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
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-canary-stage" {
  source = "./modules/workspace"

  name       = "workers-canary"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-canary/infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-canary-prod" {
  source = "./modules/workspace"

  name       = "workers-canary"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-canary/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-apps-prod" {
  source = "./modules/workspace"

  name       = "workers-apps-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-apps/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-apps-stage" {
  source = "./modules/workspace"

  name       = "workers-apps-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-apps/infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-deployments-prod" {
  source = "./modules/workspace"

  name       = "workers-deployments-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-deployments/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-deployments-stage" {
  source = "./modules/workspace"

  name       = "workers-deployments-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-deployments/infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
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
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
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
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-instances-prod" {
  source = "./modules/workspace"

  name       = "workers-instances-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-instances/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-instances-stage" {
  source = "./modules/workspace"

  name       = "workers-instances-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-instances/infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-installs-prod" {
  source = "./modules/workspace"

  name       = "workers-installs-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-installs/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-installs-stage" {
  source = "./modules/workspace"

  name       = "workers-installs-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-installs/infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-orgs-prod" {
  source = "./modules/workspace"

  name       = "workers-orgs-prod"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-orgs/infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "workers-orgs-stage" {
  source = "./modules/workspace"

  name       = "workers-orgs-stage"
  repo       = "powertoolsdev/mono"
  auto_apply = true
  dir        = "services/workers-orgs/infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}
