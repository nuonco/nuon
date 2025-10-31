locals {
  dns = {
    zone = "${var.pool}.${local.vars.region}.${var.account}.${local.vars.root_domain}"
  }
}

# Private hosted zone for internal services
resource "aws_route53_zone" "internal_private" {
  name = local.dns.zone

  force_destroy = true

  vpc {
    vpc_id = module.vpc.vpc_id
  }
}
