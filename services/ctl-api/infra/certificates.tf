local {
  root_domain = "${var.env}.nuon.co"
}

data "aws_route53_zone" "public" {
  name = "${var.root_domain}"
}

module "certificate" {
  source  = "terraform-aws-modules/acm/aws"
  version = "~> 4.0"

  domain_name         = "ctl.${var.root_domain}"
  zone_id             = data.aws_route53_zone.public.zone_id
  wait_for_validation = true
}
