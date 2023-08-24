module "mono" {
  source = "./modules/repository"

  name        = "mono"
  description = "Mono repo for all service code at Nuon."

  topics     = ["terraform", "helm", "go"]
  enable_ecr = false

  // NOTE: since we use `count` to create these resources in a loop, the ordering here must be preserved. Add a new repo
  // to the _end_ of the array to ensure we don't try to delete previous repos.
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
    "ctl-api",
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
