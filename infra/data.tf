data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "terraform_remote_state" "chart_common" {
  backend = "remote"

  config = {
    organization = "launchpaddev"
    workspaces = {
      name = "chart-common"
    }
  }
}

data "aws_ecr_repository" "ecr_repository" {
  provider = aws.infra-shared-prod
  name     = local.name
}

data "aws_vpcs" "vpcs" {
  tags = {
    environment = var.env
    pool        = local.pool
  }
}

data "aws_vpc" "vpc" {
  id = data.aws_vpcs.vpcs.ids[0]
}

data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.vpc.id]
  }

  tags = {
    environment = var.env
    pool        = local.pool
    # HACK: we should actually add a tag on the subnets for public/private
    "karpenter.sh/discovery" = "${var.env}-${local.pool}"
  }
}

data "aws_route53_zone" "public" {
  name = "${var.env}.${local.vars.public_root_domain}"
}

data "aws_route53_zone" "private" {
  # HACK: this sucks. there's not a way to query just by tags or whatever
  name   = "${local.pool}.${local.vars.region}.${var.env}.${local.vars.root_domain}"
  vpc_id = data.aws_vpcs.vpcs.ids[0]
  tags = {
    environment = var.env
    pool        = local.pool
  }
}

# NOTE(jdt): This isn't ideal but more elegant than hardcoding in CI
data "tfe_outputs" "infra-eks-nuon" {
  organization = local.terraform_organization
  workspace    = "infra-eks-${var.env}-nuon"
}
