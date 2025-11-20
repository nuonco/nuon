# Manage lifecycle policies for replicated ECR repositories

locals {
  all_extra_ecr_repos = flatten([
    module.mono.extra_ecr_repo_names,
    module.horizon.extra_ecr_repo_names,
  ])
}

resource "aws_ecr_lifecycle_policy" "replica_us_east_2" {
  for_each = toset(local.all_extra_ecr_repos)

  provider   = aws.us-east-2
  repository = each.value
  policy     = module.mono.lifecycle_policy
}
