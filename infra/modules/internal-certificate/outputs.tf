output "domain" {
  value = local.fqdn
}

output "acm_certificate_arn" {
  value = module.certificate.acm_certificate_arn
}
