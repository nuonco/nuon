resource "aws_route53_record" "bairesdev_corp" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "corp"
  type    = "CNAME"
  ttl     = 300
  records = [
    "qey1llbu.mycorpprovider.net"
  ]
}

resource "aws_route53_record" "bairesdev_google" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "google._domainkey.corp"
  type    = "CNAME"
  ttl     = 300
  records = [
    "google._domainkey.qey1llbu.mycorpprovider.net"
  ]
}

resource "aws_route53_record" "bairesdev_dmarc" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_dmarc.corp"
  type    = "CNAME"
  ttl     = 300
  records = [
    "_dmarc.mycorpprovider.net"
  ]
}

resource "aws_route53_record" "bairesdev_1024" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "1024._domainkey.corp"
  type    = "CNAME"
  ttl     = 300
  records = [
    "1024._domainkey.mycorpprovider.net"
  ]
}
