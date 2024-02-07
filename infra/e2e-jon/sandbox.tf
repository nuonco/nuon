locals {
  name = "e2e-jon"

  sandboxes_repo = "nuonco/sandboxes"
  sandboxes_branch = "main"
}

module "aws-eks-sandbox" {
  source = "./e2e"
  providers = {
    nuon = nuon.sandbox
  }

  app_name = "${local.name}-aws-eks-sandbox"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir = "aws-eks"

  east_1_count = 5
  east_2_count = 5
  west_2_count = 5

  install_role_arn = module.eks_access.iam_role_arn
}

module "aws-eks-byo-vpc-sandbox" {
  source = "./e2e"
  providers = {
    nuon = nuon.sandbox
  }

  app_name = "${local.name}-byo-vpc-sandbox"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir = "aws-eks-byovpc"

  east_1_count = 5
  east_2_count = 5
  west_2_count = 5

  install_role_arn = module.eks_access.iam_role_arn
}
