resource "aws_route53_record" "labs" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "labs"
  type    = "CNAME"
  ttl     = 300
  records = [
    "cname.vercel-dns.com."
  ]
}
