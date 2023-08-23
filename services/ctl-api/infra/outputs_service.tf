output "certificate_arn" {
  value = module.certificate.acm_certificate_arn
}

output "public_domain" {
  value = "ctl.${local.root_domain}"
}

output "internal_domain" {
  value = "ctl.${data.aws_route53_zone.private.name}"
}
