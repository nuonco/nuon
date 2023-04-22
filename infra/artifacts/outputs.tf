output "gh_role_arn" {
  value       = module.github_actions.iam_role_arn
  description = "github role"
}

output "region" {
  value       = local.aws_settings.region
}

output "bucket" {
  value = {
    name   = module.bucket.s3_bucket_id
    arn    = module.bucket.s3_bucket_arn
    region = module.bucket.s3_bucket_region
  }
}
