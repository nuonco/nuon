// wiki alias
resource "aws_route53_record" "wiki" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "wiki"
  type    = "CNAME"
  ttl     = 300
  records = [
    "internal-k8s-wiki-wikiprod-5a42d3fb8d-2119730064.us-west-2.elb.amazonaws.com"
  ]
}
