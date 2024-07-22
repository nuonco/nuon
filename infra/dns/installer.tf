resource "aws_route53_record" "installers-stage-txt-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "stage.installers"
  type    = "CNAME"
  ttl     = 300
  records = [
    "cname.vercel-dns.com."
  ]
}
