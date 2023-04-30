module "apks" {
  source = "./modules/repository"

  name        = "apks"
  description = "repo for building apks used in our images"
  topics      = ["terraform"]
}

module "chart-common" {
  source = "./modules/repository"

  name        = "chart-common"
  description = "repo for common charts"
  topics      = ["terraform", "helm"]
}

module "ci-images" {
  source = "./modules/repository"

  name        = "ci-images"
  description = "repo for ci specific container images"
  topics      = ["terraform"]
}

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

  name        = "demo"
  enable_ecr  = false
  description = "Demo repo for Nuon."
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
    "pkg ✅",
    "infra ✅",
    "protos ✅",

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

module "sandboxes" {
  source = "./modules/repository"

  name        = "sandboxes"
  description = "terraform modules for sandbox creation"
  topics      = ["terraform"]
}

module "shared_configs" {
  source = "./modules/repository"

  name        = "shared-configs"
  description = "shared configuration files"
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

module "code-jonmorehouse" {
  source = "./modules/repository"

  name                     = "code-jonmorehouse"
  description              = "personal workspace for @jonmorehouse"
  enable_ecr               = false
  enable_prod_environment  = false
  enable_stage_environment = false

  topics = ["personal-workspace"]
}
