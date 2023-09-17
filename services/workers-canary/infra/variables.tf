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

variable "api_url" {
  type        = string
  description = "api url"
}

variable "api_token" {
  type        = string
  description = "api url"
}
