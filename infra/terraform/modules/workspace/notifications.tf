resource "tfe_notification_configuration" "slack-alerts" {
  name             = "${var.name}-slack-alerts"
  enabled          = true
  destination_type = "slack"
  triggers         = ["run:errored", "run:needs_attention", ]
  url              = var.slack_notifications_webhook_url
  workspace_id     = tfe_workspace.workspace.id
}
