resource "tfe_run_trigger" "triggers" {
  for_each = toset(var.trigger_workspaces)

  workspace_id  = tfe_workspace.workspace.id
  sourceable_id = each.key
}
