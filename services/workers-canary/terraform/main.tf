locals {
  name = "seed"

  sandboxes_repo = "nuonco/sandboxes"
  sandboxes_branch = "main"
  aws_regions = ["us-east-1", "us-east-2", "us-west-1", "us-west-2", "eu-west-1", "eu-west-2"]
}


resource "random_shuffle" "aws_eks_regions" {
  input        = local.aws_regions
  result_count = var.install_count
}

module "aws-eks" {
  source = "./e2e"

  app_name = "${local.name}-aws-eks"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir = "aws-eks"

  install_prefix = "aws-eks-"

  aws = [
    {
      iam_role_arn = var.aws_eks_iam_role_arn
      regions = random_shuffle.aws_eks_regions.result
    }
  ]
}

resource "random_shuffle" "aws_ecs_regions" {
  input        = local.aws_regions
  result_count = var.install_count
}

module "aws-ecs" {
  source = "./e2e"

  app_name = "${local.name}-aws-ecs"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir = "aws-ecs"

  install_prefix = "aws-eks-"
  install_count = var.install_count
  app_runner_type = "aws-ecs"
  aws = [
    {
      iam_role_arn = var.aws_ecs_iam_role_arn
      regions = random_shuffle.aws_ecs_regions.result
    }
  ]
}
