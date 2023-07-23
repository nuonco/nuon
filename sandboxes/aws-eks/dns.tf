resource "aws_route53_zone" "internal" {
  name = var.internal_root_domain

  force_destroy = true
  vpc {
    vpc_id = module.vpc.vpc_id
  }
}

resource "aws_route53_zone" "public" {
  name = var.public_root_domain

  force_destroy = true
}
