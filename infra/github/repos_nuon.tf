module "quickstart-nuon" {
  source          = "./modules/repository"
  name            = "quickstart"
  description     = "A simple example project to easily get up and running with Nuon."
  required_checks = []
  is_public       = true
  owning_team_id  = github_team.nuon.id

  providers = {
    github = github.nuon
  }
}

module "nuonco" {
  source          = "./modules/repository"
  name            = "nuonco"
  description     = "Run your app in your customer's cloud with nuon.co"
  required_checks = []
  is_public       = true
  owning_team_id  = github_team.nuon.id

  providers = {
    github = github.nuon
  }
}
