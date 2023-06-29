locals {
  dev_public_repos = [
    "dev-public",
    "dev-waypoint-plugin-helm",
    "dev-waypoint-plugin-noop",
    "dev-waypoint-plugin-oci",
    "dev-waypoint-plugin-oci-sync",
    "dev-waypoint-plugin-terraform",
  ]
}

module "dev_public_repos" {
  source = "../modules/public-ecr"
  count  = length(local.dev_public_repos)

  name        = element(local.dev_public_repos, count.index)
  region      = local.aws_settings.public_region
  description = "ECR repo for development"
  about       = "ECR repo for pushing development containers that need to be public, such as for testing plugins."
  tags        = {}

  providers = {
    aws = aws.public
  }
}
