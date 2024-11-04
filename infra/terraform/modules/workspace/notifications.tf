resource "tfe_notification_configuration" "slack-alerts" {
  count            = var.slack_notifications_webhook_url != "" ? 1 : 0
  name             = "${var.name}-slack-alerts"
  enabled          = true
  destination_type = "slack"
  triggers         = ["run:errored", "run:needs_attention", ]
  url              = var.slack_notifications_webhook_url
  workspace_id     = tfe_workspace.workspace.id
}

resource "tfe_notification_configuration" "pagerduty-incidents" {
  count            = var.pagerduty_email_address != "" ? 1 : 0
  name             = "${var.name}-pagerduty-alerts"
  enabled          = true
  destination_type = "email"
  triggers         = ["run:errored"]
  email_addresses  = [var.pagerduty_email_address]
  workspace_id     = tfe_workspace.workspace.id
}
