output "canary_slack_webhooks_url" {
  value = var.slack_webhook_url_canary_bots
}

output "canary_install_iam_role_arn" {
  value = module.install_access.iam_role_arn
}

output "api_url" {
  value = module.install_access.iam_role_arn
}

output "api_token" {
  value = module.install_access.iam_role_arn
}
