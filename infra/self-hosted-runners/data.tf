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

locals {
  template_vars = {
    "aws_account" : var.env,
    "env" : var.env
  }
  default_vars = templatefile("vars/defaults.yaml", local.template_vars)
  env_vars     = templatefile("vars/${var.env}.yaml", local.template_vars)
}

data "utils_deep_merge_yaml" "vars" {
  input = [
    local.default_vars,
    local.env_vars,
  ]
}

data "tfe_outputs" "infra-eks-nuon" {
  organization = local.terraform_organization
  workspace    = "infra-eks-${var.env}-${local.vars.pool}"
}