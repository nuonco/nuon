// naked app dns
resource "aws_route53_record" "app" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "app"
  type    = "CNAME"
  ttl     = 300
  records = [
    "k8s-dashboar-dashboar-e338a5f879-827253081.us-west-2.elb.amazonaws.com"
  ]
}
