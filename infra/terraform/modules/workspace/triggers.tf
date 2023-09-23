resource "tfe_run_trigger" "triggers" {
  for_each = toset(var.triggered_by)

  workspace_id  = tfe_workspace.workspace.id
  sourceable_id = each.key
}
