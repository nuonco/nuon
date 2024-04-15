locals {
  root_domain = "${var.env}.${local.vars.public_root_domain}"
}

data "aws_route53_zone" "public" {
  name = local.root_domain
}

module "certificate" {
  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  domain_name         = local.domain
  zone_id             = data.aws_route53_zone.public.zone_id
  wait_for_validation = true
}
