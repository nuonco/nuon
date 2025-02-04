// status page dns
resource "aws_route53_record" "status" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "status"
  type    = "CNAME"
  ttl     = 300
  records = [
    "110422980830157.hostedstatus.com"
  ]
}
