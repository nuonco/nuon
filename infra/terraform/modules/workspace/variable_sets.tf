data "tfe_variable_set" "variable_sets" {
  count = length(var.variable_sets)

  name         = element(var.variable_sets, count.index)
  organization = data.tfe_organization.main.name
}

resource "tfe_workspace_variable_set" "variable_sets" {
  count           = length(var.variable_sets)
  workspace_id    = tfe_workspace.workspace.id
  variable_set_id = data.tfe_variable_set.variable_sets[count.index].id
}
