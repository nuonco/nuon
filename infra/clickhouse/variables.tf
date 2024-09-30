locals {
  name                   = "clickhouse"
  terraform_organization = "nuonco"
  zone                   = nonsensitive(data.tfe_outputs.infra-eks-nuon.values.private_zone)

  # clickhouse installation manifest values
  hosts              = local.vars.hosts
  replicas           = local.vars.replicas
  shards             = local.vars.shards
  availability_zones = local.vars.availability_zones

  data_volume_storage = local.vars.data_volume_storage
  image_tag           = local.vars.image_tag

  logLevel            = local.vars.logLevel

  tables              = local.vars.tables

  tags = {
    environment = var.env
    pool        = local.vars.pool
    terraform   = "infra-clickhouse-${var.env}"
  }

  vars = yamldecode(data.utils_deep_merge_yaml.vars.output)
}

variable "env" {
  type        = string
  description = "The environment to use"
}
