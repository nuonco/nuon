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
  validation_method = "DNS"
  create_route53_records = false
  validation_record_fqdns = module.root-certificate-validation[0].validation_route53_record_fqdns
}

module "root-certificate-validation" {
  count = var.use_root_domain ? 1 : 0

  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  create_certificate = false
  create_route53_records_only               = true
  validation_method = "DNS"

  zone_id                                   = data.aws_route53_zone.root.zone_id
  acm_certificate_domain_validation_options = module.root-certificate[0].acm_certificate_domain_validation_options
  distinct_domain_names = module.root-certificate[0].distinct_domain_names

  providers = {
    aws = aws.root
  }
}
