locals {
  root_fqdn = "${var.subdomain}.${var.root_domain}"
}

data "aws_route53_zone" "root" {
  name = var.root_domain

  provider = aws.root
}

module "root-certificate" {
  count = var.use_root_domain ? 1 : 0

  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  domain_name            = local.root_fqdn
  wait_for_validation    = true
  create_route53_records = false
}

module "root-certificate-validation" {
  count = var.use_root_domain ? 1 : 0

  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  domain_name                               = local.root_fqdn
  zone_id                                   = data.aws_route53_zone.root.zone_id
  wait_for_validation                       = true
  create_route53_records_only               = true
  acm_certificate_domain_validation_options = module.root-certificate[0].acm_certificate_domain_validation_options
  distinct_domain_names                     = [local.root_fqdn]

  providers = {
    aws = aws.root
  }
}
