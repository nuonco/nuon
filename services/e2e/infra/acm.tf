module "certificate" {
  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  domain_name         = "nlb.${var.public_domain}"
  zone_id             = var.public_domain_zone_id
  wait_for_validation = true
}
