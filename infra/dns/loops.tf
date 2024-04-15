resource "aws_route53_record" "mail-mx" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "envelope.mail"
  type    = "MX"
  ttl     = 300
  records = [
    "10 feedback-smtp.us-east-1.amazonses.com."
  ]
}

resource "aws_route53_record" "mail-spf" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "envelope.mail"
  type    = "TXT"
  ttl     = 300
  records = [
    "v=spf1 include:amazonses.com ~all",
  ]
}

resource "aws_route53_record" "mail-dmarc" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_dmarc.mail"
  type    = "TXT"
  ttl     = 300
  records = [
    "v=DMARC1; p=none"
  ]
}

resource "aws_route53_record" "mail-dkim-1" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "23q44okgg5ynannbvi6zraujhpizlurt._domainkey.mail"
  type    = "CNAME"
  ttl     = 300
  records = [
    "23q44okgg5ynannbvi6zraujhpizlurt.dkim.amazonses.com."
  ]
}

resource "aws_route53_record" "mail-dkim-2" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "i74k3fvx3r3ydxujlyeceog7v4yjeku7._domainkey.mail"
  type    = "CNAME"
  ttl     = 300
  records = [
    "i74k3fvx3r3ydxujlyeceog7v4yjeku7.dkim.amazonses.com."
  ]
}

resource "aws_route53_record" "mail-dkim-3" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "zb2fafxjcljatoo3gggsda7qir4qzbrc._domainkey.mail"
  type    = "CNAME"
  ttl     = 300
  records = [
    "zb2fafxjcljatoo3gggsda7qir4qzbrc.dkim.amazonses.com."
  ]
}
