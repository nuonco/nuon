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
  count            = var.pagerduty_events_api_url != "" ? 1 : 0
  name             = "${var.name}-pagerduty-alerts"
  enabled          = true
  destination_type = "generic"
  triggers         = ["run:errored"]
  url              = var.pagerduty_events_api_url
  token            = var.pagerduty_events_api_token
  workspace_id     = tfe_workspace.workspace.id
}
