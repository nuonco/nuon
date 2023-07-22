# the nuon.run domain is used to create install names in both stage and prod.
resource "aws_route53_zone" "nuon-run" {
  provider = aws.public
  name     = "nuon.run"
}

resource "aws_route53_zone" "nuon-run-stage" {
  provider = aws.public
  name     = "stage.nuon.run"
}

resource "aws_route53_record" "nuon-run-stage" {
  provider = aws.public

  zone_id = aws_route53_zone.nuon-run.zone_id
  name    = "stage"
  type    = "NS"
  ttl     = 3600
  records = aws_route53_zone.nuon-run-stage.name_servers
}
