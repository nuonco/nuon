# The actual repo being created
resource "github_repository" "main" {
  lifecycle {
    prevent_destroy = true
  }

  name        = var.name
  description = var.description

  // if is_public is set, then this will be public, if is_private is set, then it will be private, defaulting to
  // "internal"
  visibility   = var.is_public ? "public" : var.is_private ? "private" : "internal"
  archived     = var.archived
  has_issues   = true  # we used to turn off issues for archived repos. that causes issues so don't
  has_projects = false # we use org projects, not older projects v1
  has_wiki     = false # we use notion
  auto_init    = true

  # merge commits are messy
  # we don't want to see 100 "test","crap","fix", "etc" commits on main
  allow_merge_commit = false
  # squash commits are cleaner and are "verified"
  # i.e. github will sign the commit for us
  # devs sign their commits, gh signs the squash, we are "verified" end-to-end
  allow_squash_merge = true
  # rebase can also be messy and cannot be signed by GH
  allow_rebase_merge = false
  # if CI and other requirements are met, it's mergable.
  # if that's ever not the case, we should update CI / requirements, not turn this off
  allow_auto_merge = true
  # this prevents an accumulation of branches - it's good repo hygiene
  delete_branch_on_merge = true

  squash_merge_commit_title = "PR_TITLE"

  vulnerability_alerts = !var.archived # turn of dependabot alerts for archived repos

  topics = concat(["managed-by-terraform"], var.topics)
}

resource "github_team_repository" "owner" {
  repository = github_repository.main.name

  # if a team isn't given then "team" will "own" the repo and have push permission
  team_id    = var.owning_team_id
  permission = "push"
}
