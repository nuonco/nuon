// wiki alias
resource "aws_route53_record" "wiki" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "wiki"
  type    = "CNAME"
  ttl     = 300
  records = [
    "internal-k8s-wiki-wiki-e4a6410736-714817722.us-west-2.elb.amazonaws.com"
  ]
}
