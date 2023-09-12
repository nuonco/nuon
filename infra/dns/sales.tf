// point docs domain at gitbook
resource "aws_route53_record" "sales" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "sales"
  type    = "CNAME"
  ttl     = 300
  records = [
    "cname.super.so"
  ]
}
