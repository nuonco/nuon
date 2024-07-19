// dev.installer-hosted alias
// ns records for wildcard subdomain
// txt records are in ./txt.tf
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

resource "aws_route53_record" "ns_wildcard_dev_installer_hosted" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_acme-challenge"
  type    = "NS"
  ttl     = 300
  records = [
    "ns1.vercel-dns.com.",
    "ns2.vercel-dns.com."
  ]
}
