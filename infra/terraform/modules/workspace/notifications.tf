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
  count            = var.pagerduty_service_account_id != "" ? 1 : 0
  name             = "${var.name}-pagerduty-alerts"
  enabled          = true
  destination_type = "email"
  triggers         = ["run:errored"]
  email_user_ids   = [var.pagerduty_service_account_id]
  workspace_id     = tfe_workspace.workspace.id
}

resource "tfe_notification_configuration" "datadog-terraform-run-errors" {
  count            = var.datadog_terraform_run_error_email != "" ? 1 : 0
  name             = "${var.name}-datadog-terraform-run-errors"
  enabled          = true
  destination_type = "email"
  triggers         = ["run:errored"]
  email_addresses  = [var.datadog_terraform_run_error_email]
  workspace_id     = tfe_workspace.workspace.id
}
