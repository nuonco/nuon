// api alias
resource "aws_route53_record" "dev_installer_hosted" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "dev.installer-hosted"
  type    = "CNAME"
  ttl     = 300
  records = [
    "cname.vercel-dns.com."
  ]
}

resource "aws_route53_record" "wildcard_dev_installer_hosted" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "*.dev.installer-hosted"
  type    = "CNAME"
  ttl     = 300
  records = [
    "cname.vercel-dns.com."
  ]
}
