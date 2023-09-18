resource "tfe_variable" "vars" {
  for_each = var.vars

  key          = each.key
  value        = each.value
  category     = "terraform"
  workspace_id = tfe_workspace.workspace.id
  description  = "managed by terraform (via infra-terraform)."
}

resource "tfe_variable" "env_vars" {
  for_each = var.env_vars

  key          = each.key
  value        = each.value
  category     = "env"
  workspace_id = tfe_workspace.workspace.id
  description  = "managed by terraform (via infra-terraform)."
}
