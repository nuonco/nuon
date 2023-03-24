resource "github_repository_environment" "prod" {
  count = var.enable_prod_environment ? 1 : 0

  environment = "prod"
  repository  = github_repository.main.name

  deployment_branch_policy {
    protected_branches     = true
    custom_branch_policies = false
  }
  wait_timer = var.prod_wait_timer
}

resource "github_repository_environment" "stage" {
  count = var.enable_prod_environment ? 1 : 0

  environment = "stage"
  repository  = github_repository.main.name

  deployment_branch_policy {
    protected_branches     = true
    custom_branch_policies = false
  }
}
