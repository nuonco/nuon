// naked app dns
#resource "aws_route53_record" "app" {
  #zone_id = aws_route53_zone.main.zone_id
  #name    = "app"
  #type    = "CNAME"
  #ttl     = 300
  #records = [
    #"app.${data.aws_route53_zone.prod.name}"
  #]
#}
