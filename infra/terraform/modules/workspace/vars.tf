resource "tfe_variable" "vars" {
  for_each = var.vars

  key          = each.key
  value        = each.value
  category     = "terraform"
  workspace_id = tfe_workspace.workspace.id
  description  = "managed by terraform (via infra-terraform)."
}
