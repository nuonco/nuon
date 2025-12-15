resource "aws_route53_zone" "awesomebyoc_dns" {
  provider = aws.prod

  name = "awesomebyoc.com"
}

resource "aws_route53_record" "awesomebyoc_dns_root" {
  provider = aws.prod

  zone_id = aws_route53_zone.awesomebyoc_dns.zone_id
  name    = ""
  type    = "A"
  ttl     = 300
  records = [
    "216.150.1.1"
  ]
}

resource "aws_route53_record" "awesomebyoc_dns_www" {
  provider = aws.prod

  zone_id = aws_route53_zone.awesomebyoc_dns.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = 300
  records = [
    "20743dd4e791271c.vercel-dns-016.com."
  ]
}
