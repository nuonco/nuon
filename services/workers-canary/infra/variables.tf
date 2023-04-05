locals {
  name = "workers-canary"
}

variable "env" {
  description = "env"
}

variable "canary_slack_webhooks_url" {
  description = "slack webhook url for canary channel"
}
