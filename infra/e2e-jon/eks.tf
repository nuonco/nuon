module "aws-eks" {
  source = "./e2e"

  app_name = "${local.name}-aws-eks"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir = "aws-eks"

  east_1_count = 1
  east_2_count = 0
  west_2_count = 0

  install_role_arn = module.eks_access.iam_role_arn
}

module "aws-eks-byo-vpc" {
  source = "./e2e"

  app_name = "${local.name}-aws-byo-vpc"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir = "aws-eks-byovpc"

  east_1_count = 0
  east_2_count = 0
  west_2_count = 0

  install_role_arn = module.eks_access.iam_role_arn
}
