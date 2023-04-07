locals {
  name = "workers-canary"
}

variable "env" {
  type        = string
  description = "env"
}

variable "canary_slack_webhooks_url" {
  type        = string
  description = "slack webhook url for canary channel"
}
