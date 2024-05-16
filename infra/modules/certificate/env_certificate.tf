locals {
  env_fqdn = "${var.subdomain}.${var.env}.${var.root_domain}"
}

data "aws_route53_zone" "env" {
  name = "${var.env}.${var.root_domain}"
}

module "env-certificate" {
  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  domain_name         = local.env_fqdn
  zone_id             = data.aws_route53_zone.env.zone_id
  wait_for_validation = true
}

