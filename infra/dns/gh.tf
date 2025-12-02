// gh.nuon.co configs for github pages
resource "aws_route53_record" "gh_txt" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "_github-pages-challenge-nuonco.gh"
  type    = "TXT"
  ttl     = 300
  records = [
    "695830077d79f7b262b0083c146af8"
  ]
}
