module "apks" {
  source = "./modules/workspace"

  name          = "apks"
  repo          = "powertoolsdev/apks"
  auto_apply    = false
  dir           = "infra"
  variable_sets = ["aws-environment-credentials"]
  project_id    = tfe_project.infra.id

  slack_notifications_webhook_url = var.default_slack_notifications_webhook_url
  allowed_remote_state_workspaces = []
}
