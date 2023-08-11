output "certificate_arn" {
  value = module.certificate.acm_certificate_arn
}

output "certificate_domains" {
  value = module.certificate.distinct_domain_names
}
