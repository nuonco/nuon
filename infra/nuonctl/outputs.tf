output "bucket_name" {
  value = module.bucket.s3_bucket_id
}

output "bucket_arn" {
  value = module.bucket.s3_bucket_arn
}

output "bucket_region" {
  value = local.aws_settings.region
}

output "account_locks_bucket_name" {
  value = module.account_locks_bucket.s3_bucket_id
}

output "account_locks_bucket_arn" {
  value = module.account_locks_bucket.s3_bucket_arn
}
