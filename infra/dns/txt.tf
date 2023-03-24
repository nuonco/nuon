# many different tools require @ / top level txt records, adding them to individual resources will cause a race
# condition.
resource "aws_route53_record" "txt-records" {
  zone_id = aws_route53_zone.main.zone_id
  name    = ""
  type    = "TXT"
  ttl     = 300
  records = [
    # google site verification
    "google-site-verification=g5klbAXQLq5-lg-x12cHhKHSgVyGxHWBhtBRX1pJx-Q",

    # webflow
    "proxy-ssl.webflow.com"
  ]
}
