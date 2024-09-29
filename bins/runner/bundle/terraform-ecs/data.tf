locals {
  subnets_private_tag_key   = "visibility"
  subnets_private_tag_value = "private"

  private_subnet_ids = data.aws_subnets.private.ids
}

data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }

  filter {
    name   = "tag:${local.subnets_private_tag_key}"
    values = [local.subnets_private_tag_value]
  }
}
