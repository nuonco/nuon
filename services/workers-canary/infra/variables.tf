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

variable "azure_aks_subscription_id" {
  type        = string
  description = "Azure AKS subscription id"
}

variable "azure_aks_client_id" {
  type        = string
  description = "Azure AKS client id"
}

variable "azure_aks_tenant_id" {
  type        = string
  description = "Azure AKS tenant id"
}

variable "azure_aks_client_secret" {
  type        = string
  description = "Azure AKS client secret"
}
