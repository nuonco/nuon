resource "aws_route53_zone" "main" {
  name = var.root_domain
}


data "aws_route53_zone" "stage" {
  provider = aws.stage

  name = "stage.${var.root_domain}"
}

data "aws_route53_zone" "prod" {
  provider = aws.prod

  name = "prod.${var.root_domain}"
}
