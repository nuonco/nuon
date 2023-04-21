locals {
  name = "workers-canary"
}

variable "env" {
  type        = string
  description = "env"
}

variable "slack_webhook_url_canary_bots" {
  type        = string
  description = "slack webhook url for canary channel"
}
