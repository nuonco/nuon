resource "nuon_app" "main" {
  name = var.app_name
  description = "e2e managed app"
  display_name = "E2E managed app"
  slack_webhook_url = "slack webhook url"
}
