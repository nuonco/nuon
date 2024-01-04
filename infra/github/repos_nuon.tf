module "quickstart-nuon" {
  source           = "./modules/repository"
  name             = "quickstart"
  description      = "Create a fully managed version of your app that runs in your customer’s cloud account."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

# This had to start as a fork of nuonco/quickstart, so we could test
# using the quickstart in the same way a vendor would.
# Leaving this here as a record of how this repo was created.
import {
  to = module.quickstart_test_nuon.github_repository.main
  id = "quickstart-test"
}

module "quickstart_test_nuon" {
  source           = "./modules/repository"
  name             = "quickstart-test"
  description      = "Repo for testing the quickstart"
  required_checks  = []
  is_public        = true
  is_fork          = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuonco" {
  source           = "./modules/repository"
  name             = ".github"
  description      = "Run your app in your customer's cloud with nuon.co"
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "terraform-provider-nuon" {
  source           = "./modules/repository"
  name             = "terraform-provider-nuon"
  description      = "A Terraform provider for managing applications in Nuon."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-go" {
  source           = "./modules/repository"
  name             = "nuon-go"
  description      = "An SDK for interacting with the Nuon platform."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-elixir" {
  source           = "./modules/repository"
  name             = "nuon-elixir"
  description      = "An SDK for integrating with Nuon from Elixir."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-sandboxes" {
  source           = "./modules/repository"
  name             = "sandboxes"
  description      = "Builtin sandboxes for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-actions-build" {
  source           = "./modules/repository"
  name             = "actions-build"
  description      = "Action for building a Nuon component."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-actions-release" {
  source           = "./modules/repository"
  name             = "actions-release"
  description      = "Action for releasing a Nuon build."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}
