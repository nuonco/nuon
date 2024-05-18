output "certificate_arn" {
  value = module.certificate.acm_certificate_arn
}

output "domain" {
  value = module.certificate.domain
}
