output "bucket_name" {
  description = "bucket name"
  value       = local.bucket_name
}

output "runner_dev" {
  description = "runner dev IAM role arn"
  value = module.runner_dev.iam_role_arn
}
