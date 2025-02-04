// api alias
resource "aws_route53_record" "api" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api"
  type    = "CNAME"
  ttl     = 300
  records = [
    "k8s-ctlapi-ctlapiap-2cf8ef3435-810549891.us-west-2.elb.amazonaws.com"
  ]
}
