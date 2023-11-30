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

    # sender policy framework to allow gmail emails
    "v=spf1 a mx include:_spf.google.com ~all",

    # webflow
    "proxy-ssl.webflow.com"
  ]
}
