// api alias
resource "aws_route53_record" "api" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api"
  type    = "CNAME"
  ttl     = 300
  records = [
    "k8s-ctlapi-ctlapipr-db27047e57-2026103366.us-west-2.elb.amazonaws.com"
  ]
}
