locals {
  fqdn = "${var.subdomain}.${var.domain}"
}

data "aws_route53_zone" "zone" {
  name         = var.domain
  private_zone = true
  provider     = aws.default
}

module "certificate" {
  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  domain_name             = local.fqdn
  validation_method       = "DNS"
  create_route53_records  = false
  validation_record_fqdns = module.certificate-validation.validation_route53_record_fqdns
}

module "certificate-validation" {
  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  create_certificate          = false
  create_route53_records_only = true
  validation_method           = "DNS"

  zone_id                                   = data.aws_route53_zone.zone.zone_id
  acm_certificate_domain_validation_options = module.certificate.acm_certificate_domain_validation_options
  distinct_domain_names                     = module.certificate.distinct_domain_names

  providers = {
    aws = aws.default
  }
}
