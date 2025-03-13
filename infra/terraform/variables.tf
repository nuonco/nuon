variable "default_slack_notifications_webhook_url" {
  type        = string
  description = "default webhook URL for slack notifications"
}

variable "pagerduty_email_address" {
  description = "Email address for creating Pagerduty incidents."
  type        = string
}