// naked app dns
resource "aws_route53_record" "app" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "app"
  type    = "CNAME"
  ttl     = 300
  records = [
    "k8s-dashboar-dashboar-e71d0751f3-1202434589.us-west-2.elb.amazonaws.com"
  ]
}
