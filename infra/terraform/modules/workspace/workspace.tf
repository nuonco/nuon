resource "tfe_workspace" "workspace" {
  name         = var.name
  description  = "${var.name} terraform workspace for repo ${var.repo}."
  organization = data.tfe_organization.main.name

  auto_apply        = var.auto_apply
  queue_all_runs    = false
  working_directory = var.dir
  trigger_prefixes  = compact(concat([var.dir], var.trigger_prefixes))
  terraform_version = var.terraform_version

  global_remote_state       = true
  remote_state_consumer_ids = []
  project_id                = var.project_id

  tag_names = ["managed-by:terraform", "${var.auto_apply ? "auto-applied" : "manually-applied"}"]

  dynamic "vcs_repo" {
    # only create if repo is not empty.
    # this allows for manually applied workspaces
    for_each = var.repo == "" ? [] : [1]
    content {
      identifier     = var.repo
      branch         = var.vcs_branch
      oauth_token_id = data.tfe_oauth_client.github.oauth_token_id
    }
  }
}
