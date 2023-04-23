locals {
  dns = {
    zone = "${local.vars.id}.nuon.run"
  }
}

resource "aws_route53_zone" "internal_private" {
  name = local.dns.zone

  force_destroy = true

  vpc {
    vpc_id = module.vpc.vpc_id
  }
}

