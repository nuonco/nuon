// naked app dns
resource "aws_route53_record" "app" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "app"
  type    = "CNAME"
  ttl     = 300
  records = [
    "k8s-default-dashboar-fdcb6bc9e1-2132532861.us-west-2.elb.amazonaws.com"
  ]
}

// naked api dns
resource "aws_route53_record" "api" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api"
  type    = "CNAME"
  ttl     = 300
  records = [
    "k8s-default-ctlapiap-867525b026-1007387686.us-west-2.elb.amazonaws.com"
  ]
}
