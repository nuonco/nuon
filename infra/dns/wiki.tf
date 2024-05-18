// naked wiki dns
// TODO(jm): add this once wiki is promoted
#resource "aws_route53_record" "wiki" {
  #zone_id = aws_route53_zone.main.zone_id
  #name    = "wiki"
  #type    = "CNAME"
  #ttl     = 300
  #records = [
    #"k8s-default-ctlapiap-867525b026-1007387686.us-west-2.elb.amazonaws.com"
  #]
#}
