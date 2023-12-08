locals {
  name = "seed"

  sandboxes_repo = "nuonco/sandboxes"
  sandboxes_branch = "main"
}

module "aws-eks" {
  source = "./e2e"

  app_name = "${local.name}-aws-eks"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir = "aws-byo-vpc"

  east_1_count = 5
  east_2_count = 5
  west_2_count = 5

  install_role_arn = var.install_role_arn
}

module "aws-eks-byo-vpc" {
  source = "./e2e"

  app_name = "${local.name}-aws-byo-vpc"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir = "aws-byo-vpc"

  east_1_count = 0
  east_2_count = 0
  west_2_count = 0

  install_role_arn = var.install_role_arn
}
