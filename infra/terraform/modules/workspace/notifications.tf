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

resource "tfe_notification_configuration" "datadog-oncall" {
  name             = "${var.name}-datadog-oncall-alerts"
  enabled          = true
  destination_type = "generic"
  url              = "https://http-intake.logs.us5.datadoghq.com/v1/input?dd-api-key=${var.datadog_api_key}&ddsource=terraform-cloud&service=${var.name}&ddtags=env:production"
  triggers         = ["run:errored"]
  workspace_id     = tfe_workspace.workspace.id
}
