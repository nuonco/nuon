resource "aws_route53_zone" "main" {
  name = var.root_domain
}

output "nuon_nameservers" {
  value = aws_route53_zone.main.name_servers
}

data "aws_route53_zone" "stage" {
  provider = aws.stage

  name = "stage.${var.root_domain}"
}

data "aws_route53_zone" "prod" {
  provider = aws.prod

  name = "prod.${var.root_domain}"
}

# nuon.run is our publicly facing URL service, and the domain lives in the jm@powertools.dev google domains account
# since nuon.run is managed by horizon, this zone is created in the horizon account. Neither ACM or EKS support cross
# account domains (ie: for the cert validation ACM can't create validation records in an external account, and EKS can
# only use a cert in the same account).
resource "aws_route53_zone" "nuon-run" {
  provider = aws.horizon
  name     = "nuon.run"
}

output "nuon_run_nameservers" {
  value = aws_route53_zone.nuon-run.name_servers
}
