// NOTE: most of the repos will here will eventually be deprecated and moved into the mono repo.
module "dot_github" {
  source = "./modules/repository"

  name        = ".github"
  description = "shared issues / github config"
}

module "company" {
  source = "./modules/repository"

  name            = "company"
  enable_ecr      = false
  description     = "All things Nuon 🚀"
  required_checks = []
}

module "terraform-provider-echo" {
  source = "./modules/repository"

  name            = "terraform-provider-echo"
  enable_ecr      = false
  description     = "Demo echo module for working with private terraform cloud registries."
  required_checks = []
}

module "retool-seed" {
  source = "./modules/repository"

  name            = "retool-seed"
  enable_ecr      = false
  description     = "Retool seed"
  required_checks = []
}

module "graveyard" {
  source = "./modules/repository"

  name        = "graveyard"
  description = "dead code"
}
