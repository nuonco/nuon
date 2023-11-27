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

module "terraform-provider-nuon" {
  source                          = "./modules/repository"
  name                            = "terraform-provider-nuon"
  description                     = "A Terraform provider for managing applications in Nuon."
  required_checks                 = []
  is_public                       = true
  owning_team_id                  = github_team.nuon.id
  require_code_owner_reviews      = true
  required_approving_review_count = 1

  providers = {
    github = github.nuon
  }
}

module "nuon-go" {
  source          = "./modules/repository"
  name            = "nuon-go"
  description     = "An SDK for interacting with the Nuon platform."
  required_checks = []
  is_public       = true
  owning_team_id                  = github_team.nuon.id
  require_code_owner_reviews      = true
  required_approving_review_count = 1

  providers = {
    github = github.nuon
  }
}

module "nuon-sandboxes" {
  source          = "./modules/repository"
  name            = "sandboxes"
  description     = "Builtin sandboxes for Nuon apps."
  required_checks = []
  is_public       = true
  owning_team_id                  = github_team.nuon.id
  require_code_owner_reviews      = true
  required_approving_review_count = 1

  providers = {
    github = github.nuon
  }
}

module "nuon-nuonco" {
  source          = "./modules/repository"
  name            = "nuonco"
  description     = "Create a fully managed version of your app that runs in your customer’s cloud account."
  required_checks = []
  is_public       = true
  owning_team_id                  = github_team.nuon.id
  require_code_owner_reviews      = true
  required_approving_review_count = 1

  providers = {
    github = github.nuon
  }
}
