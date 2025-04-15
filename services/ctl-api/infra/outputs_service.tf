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

output "install_templates_bucket" {
  value = local.bucket_name
}

output "install_templates_bucket_base_url" {
  value = "https://${module.bucket.s3_bucket_bucket_regional_domain_name}/"
}
