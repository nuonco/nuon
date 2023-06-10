module "eslint-config-nuon" {
  source = "./modules/repository"

  name        = "eslint-config-nuon"
  description = "eslint config for typescript projects"
}

module "dot_github" {
  source = "./modules/repository"

  name        = ".github"
  description = "shared issues / github config"
}

module "demo" {
  source = "./modules/repository"

  name            = "demo"
  enable_ecr      = false
  description     = "Demo repo for Nuon."
  required_checks = ["demo ✅"]

}

module "terraform-provider-echo" {
  source = "./modules/repository"

  name            = "terraform-provider-echo"
  enable_ecr      = false
  description     = "Demo echo module for working with private terraform cloud registries."
  required_checks = []
}

module "graveyard" {
  source = "./modules/repository"

  name        = "graveyard"
  description = "dead code"
}

module "horizon" {
  source = "./modules/repository"

  name        = "horizon"
  description = "repo for managing our horizon based url service"
  topics      = ["terraform"]

  extra_ecr_repos = ["hashicorp-horizon", "hashicorp-waypoint-hzn"]
}

module "mono" {
  source = "./modules/repository"

  name        = "mono"
  description = "Mono repo for all service code at Nuon."

  topics     = ["terraform", "helm", "go"]
  enable_ecr = false
  extra_ecr_repos = [
    //services
    "api",
    "api-gateway",
    "orgs-api",
    "workers-apps",
    "workers-canary",
    "workers-deployments",
    "workers-executors",
    "workers-installs",
    "workers-instances",
    "workers-orgs",
  ]

  required_checks = [
    // lints + tests + release + deploys for services and lib/infra.
    "services ✅",
    "artifacts ✅",
    "pkg ✅",
    "infra ✅",
    "protos ✅",
    "sandboxes ✅",

    // linting and various hygiene checks
    "pull_request ✅",
    "branch ✅",
    "lint ✅",
  ]
}

module "public-docs" {
  source = "./modules/repository"

  name        = "public-docs"
  description = "public documentation"
  topics      = []

  enable_branch_protection = false
}

module "ui" {
  source = "./modules/repository"

  name        = "ui"
  description = "github repo for our ui"
}

module "waypoint" {
  source = "./modules/repository"

  name        = "waypoint"
  description = "Our internal fork of hashicorp/waypoint."
  topics      = ["terraform"]
}

