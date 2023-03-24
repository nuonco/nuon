// point docs domain at gitbook
resource "aws_route53_record" "docs" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "docs"
  type    = "CNAME"
  ttl     = 300
  records = [
    "50ae19cad5-hosting.gitbook.io"
  ]
}

