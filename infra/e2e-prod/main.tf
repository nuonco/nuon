locals {
  prefix = "e2e-stage-"

  sandboxes_repo = "nuonco/sandboxes"
  sandboxes_branch = "main"
}

module "aws-eks" {
  source = "../services/e2e/nuon/app"

  app_name = "${local.prefix}-aws-eks"

  east_1_count = 0
  east_2_count = 0
  west_2_count = 0
}

module "aws-eks-sandbox" {
  source = "../services/e2e/nuon/app"

  app_name = "${local.prefix}-aws-eks-sandbox"

  sandbox_repo = local.sandbox_repo
  sandbox_branch = local.sandbox_branch
  sandbox_dir = "aws-byo-vpc"

  east_1_install_count = 5
  east_2_install_count = 5
  west_2_install_count = 5

  install_iam_role_arn = ""
}

module "aws-eks-byo-vpc" {
  source = "../services/e2e/nuon/app"

  app_name = "${local.prefix}-aws-byo-vpc"
  east_1_count = 0
  east_2_count = 0
  west_2_count = 0
  install_iam_role_arn = ""
}

module "aws-eks-byo-vpc" {
  source = "../services/e2e/nuon/app"

  app_name = "${local.prefix}-aws-eks-sandbox"

  sandbox_repo = local.sandbox_repo
  sandbox_branch = local.sandbox_branch
  sandbox_dir = "aws-byo-vpc"

  east_1_install_count = 5
  east_2_install_count = 5
  west_2_install_count = 5

  install_iam_role_arn = ""
}

module "aws-eks-byo-vpc-sandbox" {
  source = "../services/e2e/nuon/app"

  app_name = "${local.prefix}-aws-eks-sandbox"

  sandbox_repo = local.sandbox_repo
  sandbox_branch = local.sandbox_branch
  sandbox_dir = "aws-byo-vpc"

  east_1_install_count = 5
  east_2_install_count = 5
  west_2_install_count = 5

  install_iam_role_arn = ""
}
