resource "aws_route53_record" "installers-caa-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "installers"
  type    = "CAA"
  ttl     = 300
  records = [
    "0 issuewild \"letsencrypt.org\""
  ]
}

resource "aws_route53_record" "installers-stage-caa-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "stage.installers"
  type    = "CAA"
  ttl     = 300
  records = [
    "0 issuewild \"letsencrypt.org\""
  ]
}

resource "aws_route53_record" "installers-a-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "installers"
  type    = "A"
  ttl     = 300
  records = [
    "76.76.21.21"
  ]
}

resource "aws_route53_record" "installers-stage-a-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "stage.installers"
  type    = "A"
  ttl     = 300
  records = [
    "76.76.21.21"
  ]
}

resource "aws_route53_record" "installers-stage-ns-acme-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_acme-challenge.stage.installers"
  type    = "NS"
  ttl     = 300
  records = [
    "ns1.vercel-dns.com",
    "ns2.vercel-dns.com"
  ]
}

resource "aws_route53_record" "installers-ns-acme-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_acme-challenge.installers"
  type    = "NS"
  ttl     = 300
  records = [
    "ns1.vercel-dns.com",
    "ns2.vercel-dns.com"
  ]
}
