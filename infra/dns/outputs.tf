# for the nuon.run domains
output "nuon_run_prod" {
  value = {
    nameservers = aws_route53_zone.nuon-run.name_servers
    domain      = aws_route53_zone.nuon-run.name
    zone_id     = aws_route53_zone.nuon-run.id
  }
}

output "nuon_run_stage" {
  value = {
    nameservers = aws_route53_zone.nuon-run-stage.name_servers
    domain      = aws_route53_zone.nuon-run-stage.name
    zone_id     = aws_route53_zone.nuon-run-stage.id
  }
}

# for the nuon.co domain
output "nuon_nameservers" {
  value = aws_route53_zone.main.name_servers
}
