resource "aws_route53_record" "unifygtm_cname_links" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "links.try"
  type    = "CNAME"
  ttl     = 300
  records = [
    "unifyintent.com"
  ]
}

resource "aws_route53_record" "unifygtm_cname_acm_validation" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_47cea5c6b89dbcecea288bfcc67fb778.try.nuon.co"
  type    = "CNAME"
  ttl     = 300
  records = [
    "_0faa9b4ec282968108dd6c8f1481ca05.sdgjtdhdhz.acm-validations.aws"
  ]
}

resource "aws_route53_record" "unifygtm_mx" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "try"
  type    = "MX"
  ttl     = 300
  records = [
    "mxa.mailgun.org",
    "mxb.mailgun.org"
  ]
}

resource "aws_route53_record" "unifygtm_txt_spf" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "try"
  type    = "TXT"
  ttl     = 300
  records = [
    "v=spf1 include:mailgun.org ~all"
  ]
}

resource "aws_route53_record" "unifygtm_txt_dmarc" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_dmarc.try"
  type    = "TXT"
  ttl     = 300
  records = [
    "v=DMARC1; p=none; rua=mailto:dmarc-reports@unifygtm.com"
  ]
}

resource "aws_route53_record" "unifygtm_txt_domainkey" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "mx._domainkey.try"
  type    = "TXT"
  ttl     = 300
  records = [
    "k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDNaAHta57lg6Hs1OYQC5sKSAw7aCHZhOxuVMQvvQaTHOUjww7ez3SSIDYkn8svpiXqKs3YSTh7nUBPeAy0WNhhl/ZIc4n+j3XM5BI53f2lOXpJdeMWhB5L+EtdOTcvPbQlstdWtplfO2vCqV4DEefO6alEi6ginLrPrED1ahxGAQIDAQAB"
  ]
}
