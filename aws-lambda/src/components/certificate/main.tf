module "certificate" {
  source  = "terraform-aws-modules/acm/aws"
  version = "4.0"

  domain_name         = var.domain_name
  zone_id             = var.zone_id
  validation_method   = "DNS"
  wait_for_validation = false
}
