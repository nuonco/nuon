output "canary_slack_webhooks_url" {
  value = var.slack_webhook_url_canary_bots
}

output "canary_install_iam_role_arn" {
  value = module.eks_access.iam_role_arn
}

output "canary_ecs_iam_role_arn" {
  value = module.ecs_access.iam_role_arn
}

output "api_url" {
  value = var.api_url
}

output "internal_api_url" {
  value = var.internal_api_url
}

output "state_bucket_name" {
  value = local.bucket_name
}

output "state_bucket_region" {
  value = local.region
}

# for actually running canaries and creating installs
output "aws_eks_iam_role_arn" {
  value = module.eks_access.iam_role_arn
}

output "aws_ecs_iam_role_arn" {
  value = module.ecs_access.iam_role_arn
}

output "azure_aks_subscription_id" {
  value = var.azure_aks_subscription_id
}

output "azure_aks_tenant_id" {
  value = var.azure_aks_tenant_id
}

output "azure_aks_client_id" {
  value = var.azure_aks_client_id
}

output "azure_aks_client_id" {
  value = var.azure_aks_client_secret
}
