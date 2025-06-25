locals {
  name                   = "${var.env}-${local.vars.pool}-${local.service}"
  service                = "buildkit"
  terraform_organization = "nuonco"
  zone                   = nonsensitive(data.tfe_outputs.infra-eks-nuon.values.private_zone)

  tags = {
    environment = var.env
    service     = local.service
    pool        = local.vars.pool
    terraform   = "infra-buildkit-${var.env}"
  }

  vars = yamldecode(data.utils_deep_merge_yaml.vars.output)
}

variable "env" {
  type        = string
  description = "The environment to use"
  default     = "infra-shared-ci"
}

variable "datadog_api_key" {
  type        = string
  description = "Datadog API key"
  sensitive   = true
}

variable "datadog_app_key" {
  type        = string
  description = "Datadog app key"
  sensitive   = true
}
