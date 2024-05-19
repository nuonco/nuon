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
  description = "api url set by the standard api var set"
}

variable "internal_api_url" {
  type        = string
  description = "internal api url set by the standard api var set"
}

variable "github_install_id" {
  type        = string
  description = "github install id to add to org"
}
