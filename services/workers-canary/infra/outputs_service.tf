output "canary_slack_webhooks_url" {
  value = var.slack_webhook_url_canary_bots
}

output "canary_install_iam_role_arn" {
  value = module.install_access.iam_role_arn
}

output "api_url" {
  value = var.api_url
}

output "internal_api_url" {
  value = var.internal_api_url
}
