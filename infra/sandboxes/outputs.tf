output "gh_role_arn" {
  value = aws_iam_role.github_actions.arn
}

output "region" {
  value = local.region
}

output "sandboxes_url" {
  value = "s3://${local.sandbox_bucket_name}/sandboxes"
}

output "bucket" {
  value = {
    name   = module.bucket.s3_bucket_id
    arn    = module.bucket.s3_bucket_arn
    region = module.bucket.s3_bucket_region
  }
}

output "key" {
  value = {
    id  = aws_kms_key.sandbox_bucket.id
    arn = aws_kms_key.sandbox_bucket.arn
  }
}
