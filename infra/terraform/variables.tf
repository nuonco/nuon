variable "default_slack_notifications_webhook_url" {
  type        = string
  description = "default webhook URL for slack notifications"
}

variable "pagerduty_events_api_url" {
  description = "pagerduty events API URL for creating incidents"
  type        = string
}

variable "pagerduty_events_api_token" {
  description = "pagerduty events token for creating incidents"
  type        = string
}
