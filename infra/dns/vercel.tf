resource "aws_route53_record" "vercel" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_vercel"
  type    = "TXT"
  ttl     = 300
  records = [
    "vc-domain-verify=www.nuon.co,eb46ec2a8f783d42b56a",
    "vc-domain-verify=nuon.co,accce88f274d635e2146",
    "vc-domain-verify=docs.nuon.co,d405ffa1f9db2909d3bd"
  ]
}
