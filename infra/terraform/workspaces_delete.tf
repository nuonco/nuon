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
