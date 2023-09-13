module "mono" {
  source = "./modules/repository"

  name        = "mono"
  description = "Mono repo for all service code at Nuon."

  topics     = ["terraform", "helm", "go"]
  enable_ecr = false

  // NOTE: since we use `count` to create these resources in a loop, the ordering here must be preserved. Add a new repo
  // to the _end_ of the array to ensure we don't try to delete previous repos.
  extra_ecr_repos = local.mono_vars.ecr_repos
  required_checks = local.mono_vars.required_checks
}
