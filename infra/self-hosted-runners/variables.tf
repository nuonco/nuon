locals {
  name                   = "${var.env}-${local.vars.pool}-${local.service}"
  service                = "self-hosted-runners"
  terraform_organization = "nuonco"
  zone                   = nonsensitive(data.tfe_outputs.infra-eks-nuon.values.private_zone)

  tags = {
    environment = var.env
    service     = local.service
    pool        = local.vars.pool
    terraform   = "infra-self-hosted-runners-${var.env}"
  }

  vars = yamldecode(data.utils_deep_merge_yaml.vars.output)
}

variable "env" {
  type        = string
  description = "The environment to use"
  default     = "infra-shared-ci"
}

variable "github_app_private_key" {
  type        = string
  description = "GitHub App Private Key for runner authentication (sensitive)"
  sensitive   = true
}