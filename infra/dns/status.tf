resource "aws_route53_record" "status_mail_cname" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "em3855.status"
  type    = "CNAME"
  ttl     = 300
  records = [
    "u31181182.wl183.sendgrid.net"
  ]
}

resource "aws_route53_record" "status_dkim1" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "pdt._domainkey.status"
  type    = "CNAME"
  ttl     = 300
  records = [
    "pdt.domainkey.u31181182.wl183.sendgrid.net"
  ]
}

resource "aws_route53_record" "status_dkim2" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "pdt2._domainkey.status"
  type    = "CNAME"
  ttl     = 300
  records = [
    "pdt2.domainkey.u31181182.wl183.sendgrid.net"
  ]
}

resource "aws_route53_record" "status_tls_certificate" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_a78c5e0f2a62a49426862061cb1622ba.status"
  type    = "CNAME"
  ttl     = 300
  records = [
    "_1d37c496b6d110b3ac2b2b79ef42f36b.xlfgrmvvlj.acm-validations.aws"
  ]
}

resource "aws_route53_record" "status_http_traffic" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "status"
  type    = "CNAME"
  ttl     = 300
  records = [
    "cd-4dbd1d152ff49c35341d53414002a981.hosted-status.pagerduty.com"
  ]
}
