data "github_branch" "main" {
  repository = github_repository.main.name
  branch     = "main"
}

resource "github_branch_default" "default" {
  repository = github_repository.main.name
  branch     = data.github_branch.main.branch
}

resource "github_branch_protection" "default" {
  count         = var.enable_branch_protection ? 1 : 0
  repository_id = github_repository.main.node_id

  pattern = data.github_branch.main.branch

  # rules should apply to everyone equally.
  # we _should_ have a policy for how to circumvent this under exigent circumstances
  enforce_admins = true

  # this is good repo hygiene
  required_linear_history = true

  # these will be a requirement for SLSA, etc
  allows_deletions       = false
  allows_force_pushes    = false
  require_signed_commits = true

  required_pull_request_reviews {
    dismiss_stale_reviews           = true # if this is a bottleneck, the repo should be broken up
    required_approving_review_count = var.required_approving_review_count
    require_code_owner_reviews      = var.require_code_owner_reviews
  }

  required_status_checks {
    strict = true
    # This is a generic, required check.
    # The meaning of it is left to each repo to determine
    contexts = var.required_checks
  }
}
