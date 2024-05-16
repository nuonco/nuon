// NOTE(jm): this can be deleted once we get rid of the ctl.prod.nuon.co domain
locals {
  root_domain = "${var.env}.${local.vars.root_domain}"
}

data "aws_route53_zone" "public" {
  name = local.root_domain
}

module "certificate" {
  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  domain_name         = "ctl.${local.root_domain}"
  zone_id             = data.aws_route53_zone.public.zone_id
  wait_for_validation = true
}
