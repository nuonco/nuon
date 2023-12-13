data "aws_vpc" "vpc" {
  id = var.vpc_id
}

data "aws_route53_zone" "private" {
  zone_id = var.zone_id
}
