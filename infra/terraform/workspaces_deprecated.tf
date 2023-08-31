module "ci-images" {
  source = "./modules/workspace"

  name          = "ci-images"
  repo          = "powertoolsdev/ci-images"
  auto_apply    = false
  dir           = "infra"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
}
