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

data "aws_route53_zone" "private" {
  name   = local.vars.root_domain
  vpc_id = data.aws_vpcs.vpcs.ids[0]
}

data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.vpc.id]
  }

  tags = {
    environment = var.env
    pool        = local.vars.pool
    # HACK: we should actually add a tag on the subnets for public/private
    "karpenter.sh/discovery" = "${var.env}-${local.vars.pool}"
  }
}

data "tfe_outputs" "infra-orgs" {
  organization = local.terraform_organization
  workspace    = "infra-orgs-${var.env}"
}
