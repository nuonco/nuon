resource "aws_route53_record" "vercel" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_vercel"
  type    = "TXT"
  ttl     = 300
  records = [
    "vc-domain-verify=www.nuon.co,eb46ec2a8f783d42b56a",
    "vc-domain-verify=nuon.co,accce88f274d635e2146",
    "vc-domain-verify=docs.nuon.co,d405ffa1f9db2909d3bd",
    "vc-domain-verify=stage.installers.nuon.co,f6c421d292fc664dd508",
    "vc-domain-verify=*.stage.installers.nuon.co,19c51a544f53a328fb9b",
    "vc-domain-verify=installers.nuon.co,e63f62ad658d38916834"
  ]
}
