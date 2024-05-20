module "fork-test" {
  source = "./modules/repository"

  name                     = "test-installer-fork"
  description              = "Installer."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false
  auto_init = false

  providers = {
    github = github.nuonco-shared
  }
}
