locals {
  root_fqdn = "${var.subdomain}.${var.root_domain}"
}

data "aws_route53_zone" "root" {
  name = var.root_domain

  provider = aws.root
}

module "root-certificate" {
  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  domain_name            = local.root_fqdn
  zone_id                = data.aws_route53_zone.root.zone_id
  wait_for_validation    = true
  create_route53_records = false
}

module "root-certificate-validation" {
  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  domain_name                               = var.root_domain
  zone_id                                   = data.aws_route53_zone.root.zone_id
  wait_for_validation                       = true
  create_route53_records_only               = true
  acm_certificate_domain_validation_options = module.root-certificate.acm_certificate_domain_validation_options

  providers = {
    aws = aws.root
  }
}
