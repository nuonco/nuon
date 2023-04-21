output "canary_slack_webhooks_url" {
  value = nonsensitive(var.slack_webhook_url_canary_bots)
}

output "canary_install_iam_role_arn" {
  value = module.install_access.iam_role_arn
}
