resource "aws_route53_record" "installers-caa-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "installers"
  type    = "CAA"
  ttl     = 300
  records = [
    "0 issue \"letsencrypt.org\""
  ]
}
resource "aws_route53_record" "installers-cname-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "installers"
  type    = "CNAME"
  ttl     = 300
  records = [
    "cname.vercel-dns.com."
  ]
}

resource "aws_route53_record" "installers-stage-cname-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "stage.installers"
  type    = "CNAME"
  ttl     = 300
  records = [
    "cname.vercel-dns.com."
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
