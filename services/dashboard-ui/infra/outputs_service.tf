output "certificate_arn" {
  value = module.certificate.acm_certificate_arn
}

output "public_domain" {
  value = local.domain
}
