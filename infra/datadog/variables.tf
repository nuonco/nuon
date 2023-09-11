locals {
  name                   = "datadog"
  terraform_organization = "nuonco"

  zone = nonsensitive(data.tfe_outputs.infra-eks-nuon.values.private_zone)

  tags = {
    environment = var.env
    pool        = local.vars.pool
    terraform   = "infra-datadog-${var.env}"
  }

  vars = yamldecode(data.utils_deep_merge_yaml.vars.output)
}

variable "env" {
  type        = string
  description = "The environment to use"
}

variable "datadog_api_key" {
  type        = string
  description = "The datadog api key - used by the agents."
}

variable "datadog_app_key" {
  type        = string
  description = "The datadog app key"
}
