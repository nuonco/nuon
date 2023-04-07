output "canary_slack_webhooks_url" {
  value = nonsensitive(var.canary_slack_webhooks_url)
}

output "canary_install_iam_role_arn" {
  value = module.install_access.iam_role_arn
}
