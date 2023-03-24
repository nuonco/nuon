// naked app dns
resource "aws_route53_record" "app" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "app"
  type    = "CNAME"
  ttl     = 300
  records = [
    "app.${data.aws_route53_zone.prod.name}"
  ]
}

// prod app dns
resource "aws_route53_record" "app-prod" {
  provider = aws.prod

  zone_id = data.aws_route53_zone.prod.zone_id
  name    = "app"
  type    = "CNAME"
  ttl     = 300
  records = [
    "nuon-ui.fly.dev"
  ]
}


// stage app dns
resource "aws_route53_record" "app-stage" {
  provider = aws.stage

  zone_id = data.aws_route53_zone.stage.zone_id
  name    = "app"
  type    = "CNAME"
  ttl     = 300
  records = [
    "nuon-ui-stage.fly.dev"
  ]
}
