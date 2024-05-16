output "domain" {
  value = var.use_root_domain ? local.root_fqdn : local.env_fqdn
}

output "acm_certificate_arn" {
  value = var.use_root_domain ? module.root-certificate[0].acm_certificate_arn : module.env-certificate[0].acm_certificate_arn
}
