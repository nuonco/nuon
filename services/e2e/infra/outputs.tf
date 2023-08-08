output "repo" {
  value = module.repo.all
}

output "repo_name" {
  value = module.repo.name
}

output "bucket" {
  value = {
    name = local.bucket_name
  }
}

output "bucket_name" {
  value = local.bucket_name
}

output "public_domain_certificate_arn" {
  value = module.certificate.acm_certificate_arn
}
