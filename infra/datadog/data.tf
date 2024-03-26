data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "aws_vpcs" "vpcs" {
  tags = {
    environment = var.env
    pool        = local.vars.pool
  }
}

data "aws_vpc" "vpc" {
  id = data.aws_vpcs.vpcs.ids[0]
}

locals {
  vpc = {
    id         = data.aws_vpc.vpc.id
    cidr_block = data.aws_vpc.vpc.cidr_block_associations[0].cidr_block
  }
}

data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [local.vpc.id]
  }

  tags = {
    environment = var.env
    pool        = local.vars.pool
    # HACK: we should actually add a tag on the subnets for public/private
    "karpenter.sh/discovery" = "${var.env}-${local.vars.pool}"
  }
}

data "aws_route53_zone" "private" {
  # HACK: this sucks. there's not a way to query just by tags or whatever
  name   = "${local.vars.pool}.${local.vars.region}.${var.env}.${local.vars.root_domain}"
  vpc_id = data.aws_vpcs.vpcs.ids[0]
  tags = {
    environment = var.env
    pool        = local.vars.pool
  }
}

data "utils_deep_merge_yaml" "vars" {
  input = [
    file("vars/defaults.yaml"),
    file("vars/${var.env}.yaml"),
  ]
}

locals {
  values_env_path = "${path.module}/values/${var.env}.yaml"
}

data "utils_deep_merge_yaml" "values" {
  input = [
    file("${path.module}/values/datadog.yaml"),
    fileexists(local.values_env_path) ? file(local.values_env_path) : "",
  ]
}

locals {
  values = yamldecode(data.utils_deep_merge_yaml.values.output)
}

data "tfe_outputs" "infra-eks-nuon" {
  organization = local.terraform_organization
  workspace    = "infra-eks-${var.env}-${local.vars.pool}"
}
