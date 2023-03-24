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
  repo       = "powertoolsdev/api-gateway"
  auto_apply = true
  dir        = "infra"
  vars = {
    env = "stage"
  }
  variable_sets = ["aws-environment-credentials"]

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "api-gateway-prod" {
  source = "./modules/workspace"

  name       = "api-gateway-prod"
  repo       = "powertoolsdev/api-gateway"
  auto_apply = true
  dir        = "infra"
  vars = {
    env = "prod"
  }
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
}

module "apks" {
  source = "./modules/workspace"

  name          = "apks"
  repo          = "powertoolsdev/apks"
  auto_apply    = false
  dir           = "infra"
  variable_sets = ["aws-environment-credentials"]

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  allowed_remote_state_workspaces = [module.ci-images.workspace_id, ]
}

module "chart-common" {
  source = "./modules/workspace"

  name                            = "chart-common"
  repo                            = "powertoolsdev/chart-common"
  auto_apply                      = true
  dir                             = "infra"
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

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "demo" {
  source = "./modules/workspace"

  name          = "demo"
  repo          = "powertoolsdev/demo"
  auto_apply    = true
  dir           = "terraform"
  variable_sets = ["aws-environment-credentials"]

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
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "infra-eks-horizon-main" {
  source = "./modules/workspace"

  name                            = "infra-eks-horizon-main"
  repo                            = "powertoolsdev/infra-eks-nuon"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "twingate-api-token"]
  vars = {
    account = "horizon"
    pool    = "main"
  }
}

module "infra-eks-orgs-prod-main" {
  source = "./modules/workspace"

  name                            = "infra-eks-orgs-prod-main"
  repo                            = "powertoolsdev/infra-eks-nuon"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "twingate-api-token"]
  allowed_remote_state_workspaces = [
    module.infra-orgs-prod.workspace_id,
    module.workers-orgs-prod.workspace_id,
    module.workers-installs-prod.workspace_id,
    module.workers-apps-prod.workspace_id,
    module.workers-deployments-prod.workspace_id,
    module.workers-executors-prod.workspace_id,
  module.workers-instances-prod.workspace_id]

  vars = {
    account = "orgs-prod"
    pool    = "main"
  }
}

module "infra-eks-orgs-stage-main" {
  source = "./modules/workspace"

  name                            = "infra-eks-orgs-stage-main"
  repo                            = "powertoolsdev/infra-eks-nuon"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "twingate-api-token"]
  allowed_remote_state_workspaces = [
    module.infra-orgs-stage.workspace_id,
    module.workers-orgs-stage.workspace_id,
    module.workers-installs-stage.workspace_id,
    module.workers-apps-stage.workspace_id,
    module.workers-deployments-stage.workspace_id,
    module.workers-executors-stage.workspace_id,
  module.workers-instances-stage.workspace_id]
  vars = {
    account = "orgs-stage"
    pool    = "main"
  }
}

module "infra-eks-prod-nuon" {
  source = "./modules/workspace"

  name                            = "infra-eks-prod-nuon"
  repo                            = "powertoolsdev/infra-eks-nuon"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "twingate-api-token"]
  allowed_remote_state_workspaces = ["global"]
  vars = {
    account = "prod"
    pool    = "nuon"
  }
}

module "infra-eks-stage-nuon" {
  source = "./modules/workspace"

  name                            = "infra-eks-stage-nuon"
  repo                            = "powertoolsdev/infra-eks-nuon"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "twingate-api-token"]
  allowed_remote_state_workspaces = ["global"]
  vars = {
    account = "stage"
    pool    = "nuon"
  }
}

module "infra-eks-sandbox-jtarasovic" {
  source = "./modules/workspace"

  name                            = "infra-eks-sandbox-jtarasovic"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "twingate-api-token"]
  vars = {
    account = "sandbox"
    pool    = "jtarasovic"
  }
}

module "infra-github" {
  source = "./modules/workspace"

  name                            = "infra-github"
  repo                            = "powertoolsdev/infra-github"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials", "github-admin-powertoolsdev"]
}

module "infra-grafana" {
  source = "./modules/workspace"

  name                            = "infra-grafana"
  repo                            = "powertoolsdev/infra-grafana"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = []
  allowed_remote_state_workspaces = [
    module.infra-eks-stage-nuon.workspace_id,
    module.infra-eks-prod-nuon.workspace_id,
    module.infra-eks-sandbox-jtarasovic.workspace_id,
    module.infra-eks-horizon-main.workspace_id,
    module.infra-eks-orgs-prod-main.workspace_id,
  module.infra-eks-orgs-stage-main.workspace_id]
}

module "infra-orgs-prod" {
  source = "./modules/workspace"

  name       = "infra-orgs-prod"
  repo       = "powertoolsdev/infra-orgs"
  auto_apply = false
  vars = {
    env = "prod"

    deployments_bucket_name   = "nuon-org-deployments-prod"
    installations_bucket_name = "nuon-org-installations-prod"
    orgs_bucket_name          = "nuon-orgs-prod"
    secrets_bucket_name       = "nuon-org-secrets-prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  allowed_remote_state_workspaces = [
    module.orgs-api-prod.workspace_id,
    module.workers-apps-prod.workspace_id,
    module.workers-installs-prod.workspace_id,
    module.workers-instances-prod.workspace_id,
    module.workers-executors-prod.workspace_id,
    module.workers-deployments-prod.workspace_id,
    module.workers-orgs-prod.workspace_id,
  ]
}

module "infra-orgs-stage" {
  source = "./modules/workspace"

  name       = "infra-orgs-stage"
  repo       = "powertoolsdev/infra-orgs"
  auto_apply = true
  vars = {
    env = "stage"

    deployments_bucket_name   = "nuon-org-deployments-stage"
    installations_bucket_name = "nuon-org-installations-stage"
    orgs_bucket_name          = "nuon-orgs-stage"
    secrets_bucket_name       = "nuon-org-secrets-stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  allowed_remote_state_workspaces = [
    module.orgs-api-stage.workspace_id,
    module.workers-apps-stage.workspace_id,
    module.workers-installs-stage.workspace_id,
    module.workers-instances-stage.workspace_id,
    module.workers-executors-stage.workspace_id,
    module.workers-deployments-stage.workspace_id,
    module.workers-orgs-stage.workspace_id,
  ]
}

module "infra-temporal-prod" {
  source = "./modules/workspace"

  name                            = "infra-temporal-prod"
  repo                            = "powertoolsdev/infra-temporal"
  auto_apply                      = false
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  vars = {
    env = "prod"
  }
}

module "infra-temporal-stage" {
  source = "./modules/workspace"

  name                            = "infra-temporal-stage"
  repo                            = "powertoolsdev/infra-temporal"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  vars = {
    env = "stage"
  }
}

module "infra-terraform" {
  source = "./modules/workspace"

  name                            = "infra-terraform"
  repo                            = "powertoolsdev/infra-terraform"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "nuon-dns" {
  source = "./modules/workspace"

  name                            = "nuon-dns"
  repo                            = "powertoolsdev/infra-nuon-dns"
  auto_apply                      = false
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

module "powertools" {
  source = "./modules/workspace"

  name                            = "powertools"
  repo                            = "powertoolsdev/infra-powertools"
  auto_apply                      = true
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "sandboxes" {
  source = "./modules/workspace"

  name                            = "sandboxes"
  repo                            = "powertoolsdev/sandboxes"
  auto_apply                      = true
  dir                             = "infra"
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  variable_sets                   = ["aws-environment-credentials"]
  allowed_remote_state_workspaces = [
    module.infra-orgs-prod.workspace_id,
    module.infra-orgs-stage.workspace_id,
  ]
}

module "template-go-service-prod" {
  source = "./modules/workspace"

  name       = "template-go-service-prod"
  repo       = "powertoolsdev/template-go-service"
  auto_apply = true
  dir        = "infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "template-go-service-stage" {
  source = "./modules/workspace"

  name       = "template-go-service-stage"
  repo       = "powertoolsdev/template-go-service"
  auto_apply = true
  dir        = "infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "terraform-runner-prod" {
  source = "./modules/workspace"

  name       = "terraform-runner-prod"
  repo       = "powertoolsdev/terraform-runner"
  auto_apply = false
  dir        = "infra"
  vars = {
    env = "prod"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "terraform-runner-stage" {
  source = "./modules/workspace"

  name       = "terraform-runner-stage"
  repo       = "powertoolsdev/terraform-runner"
  auto_apply = true
  dir        = "infra"
  vars = {
    env = "stage"
  }
  variable_sets                   = ["aws-environment-credentials"]
  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}

module "waypoint" {
  source = "./modules/workspace"

  name                            = "waypoint"
  repo                            = "powertoolsdev/waypoint"
  auto_apply                      = true
  dir                             = "infra"
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
