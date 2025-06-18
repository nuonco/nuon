// runner alias
resource "aws_route53_record" "runner" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "runner"
  type    = "CNAME"
  ttl     = 300
  records = [
    "k8s-ctlapi-ctlapipr-ecdf369dfb-1708754594.us-west-2.elb.amazonaws.com"
  ]
}
