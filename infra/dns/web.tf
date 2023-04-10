// dns records for web properties
resource "aws_route53_record" "www-naked" {
  zone_id = aws_route53_zone.main.zone_id
  name    = ""
  type    = "A"
  ttl     = 300
  records = [
    "75.2.70.75",
    "99.83.190.102"
  ]
}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = 300
  records = [
    "proxy-ssl.webflow.com"
  ]
}
