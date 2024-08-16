locals {
  name                   = "clickhouse"
  terraform_organization = "nuonco"
  zone                   = nonsensitive(data.tfe_outputs.infra-eks-nuon.values.private_zone)

  # clickhouse installation manifest values
  replicas            = local.vars.replicas
  shards              = local.vars.shards
  data_volume_storage = local.vars.data_volume_storage
  image_tag           = local.vars.image_tag

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
