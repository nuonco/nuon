output "certificate_arn_legacy" {
  value = module.certificate.acm_certificate_arn
}

output "public_domain_legacy" {
  value = "ctl.${local.root_domain}"
}

output "internal_domain" {
  value = "ctl.${data.aws_route53_zone.private.name}"
}

output "tfe_token" {
  value = var.tfe_token
}

output "certificate_arn" {
  value = module.cert.acm_certificate_arn
}

output "public_domain" {
  value = module.cert.domain
}

output "runner_domain" {
  value = module.runner-cert.domain
}

output "runner_certificate_arn" {
  value = module.runner-cert.acm_certificate_arn
}
