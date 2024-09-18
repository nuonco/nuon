data "twingate_groups" "engineers" {
  name = "engineers"
}

data "twingate_groups" "internal_access" {
  name = "Internal Access"
}

// this is the public dns name, created by `infra-aws`
data "aws_route53_zone" "env_root" {
  name = "${var.account}.nuon.co"
}

data "tfe_outputs" "infra-eks-nuon" {
  organization = local.terraform_organization
  workspace    = "infra-eks-${var.account}-${var.pool}"
}

data "aws_caller_identity" "current" {}
