output "runner_id" {
  value = var.runner_id
}

output "api_url" {
  value = var.api_url
}

output "api_token" {
  value = var.api_token
}

output "cluster_arn" {
  value = var.cluster_arn
}

output "install_iam_role_arn" {
  value = var.install_iam_role_arn
}

output "runner_iam_role_arn" {
  value = var.runner_iam_role_arn
}

output "vpc_id" {
  value = var.vpc_id
}

output "private_subnet_ids" {
  value = local.private_subnet_ids
}
