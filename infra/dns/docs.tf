// point docs domain at gitbook
resource "aws_route53_record" "docs" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "docs"
  type    = "CNAME"
  ttl     = 300
  records = [
    "cname.vercel-dns.com."
  ]
}

resource "aws_route53_record" "docs-txt" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_vercel"
  type    = "TXT"
  ttl     = 300
  records = [
    "vc-domain-verify=docs.nuon.co,d405ffa1f9db2909d3bd"
  ]
}
