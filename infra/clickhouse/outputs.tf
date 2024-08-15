output "clickhouse_backups_url" {
  value = "s3://${local.bucket_name}/backups"
}

output "bucket" {
  value = {
    name   = module.bucket.s3_bucket_id
    arn    = module.bucket.s3_bucket_arn
    region = module.bucket.s3_bucket_region
  }
}

