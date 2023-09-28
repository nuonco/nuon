resource "twingate_service_account" "github_actions" {
  name = "github-actions-${local.workspace_trimmed}"
}

// Key rotation using the time provider (see https://registry.terraform.io/providers/hashicorp/time/latest)
resource "time_rotating" "key_rotation" {
  rotation_days = 90
}

resource "time_static" "key_rotation" {
  rfc3339 = time_rotating.key_rotation.rfc3339
}

resource "twingate_service_account_key" "github_actions" {
  name               = "github-actions-key"
  service_account_id = twingate_service_account.github_actions.id

  lifecycle {
    replace_triggered_by = [time_static.key_rotation]
  }
}
