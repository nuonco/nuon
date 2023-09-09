data "tfe_organization" "main" {
  // NOTE(jm): our terraform cloud account was created before we changed to Nuon formally. Eventually, we should change
  // the name, but that will require a migration to prevent breaking our workspaces.
  name = "nuonco"
}
