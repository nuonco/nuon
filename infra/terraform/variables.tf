variable "default_slack_notifications_webhook_url" {
  type        = string
  description = "default webhook URL for slack notifications"
}

variable "pagerduty_email_address" {
  description = "Email address for creating Pagerduty incidents."
  type        = string
}

variable "datadog_terraform_run_error_email" {
  description = "Email address to notify on Terraform run errors in Datadog."
  type        = string
}

