// dns records for web properties
resource "aws_route53_record" "www-naked" {
  zone_id = aws_route53_zone.main.zone_id
  name    = ""
  type    = "A"
  ttl     = 300
  records = [
    "76.76.21.21",
  ]
}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = 300
  records = [
    "45828c60ae8a324a.vercel-dns-016.com"
  ]
}
