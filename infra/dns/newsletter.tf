resource "aws_route53_record" "mailchimp-verification-1" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "k2._domainkey"
  type    = "CNAME"
  ttl     = 300
  records = [
    "dkim2.mcsv.net"
  ]
}

resource "aws_route53_record" "mailchimp-verification-2" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "k3._domainkey"
  type    = "CNAME"
  ttl     = 300
  records = [
    "dkim3.mcsv.net"
  ]
}
