resource "aws_route53_record" "installers-stage-txt-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = ""
  type    = "TXT"
  ttl     = 300
  records = [
    "vc-domain-verify=stage.installers.nuon.co,f6c421d292fc664dd508"
  ]
}
