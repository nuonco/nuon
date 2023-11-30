# NOTE: temporarily disable so we can re-start authentication process
resource "aws_route53_record" "gmail_mx" {
  zone_id = aws_route53_zone.main.zone_id
  name    = ""
  type    = "MX"
  ttl     = 300
  records = [
    "1 ASPMX.L.GOOGLE.COM.",
    "5 ALT1.ASPMX.L.GOOGLE.COM.",
    "5 ALT2.ASPMX.L.GOOGLE.COM.",
    "10 ALT3.ASPMX.L.GOOGLE.COM.",
    "10 ALT4.ASPMX.L.GOOGLE.COM."
  ]
}

resource "aws_route53_record" "dmarc" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_dmarc"
  type    = "TXT"
  ttl     = 300
  records = [
    "v=DMARC1; p=none"
  ]
}
