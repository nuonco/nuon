locals {
  name                   = "waypoint"
  terraform_organization = "nuonco"

  tags = {
    environment = var.env
    pool        = local.vars.pool
    terraform   = "infra-waypoint-${var.env}"
  }

  vars = yamldecode(data.utils_deep_merge_yaml.vars.output)
}

variable "env" {
  type        = string
  description = "The environment to use"
}
