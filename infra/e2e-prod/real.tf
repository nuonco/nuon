module "aws-eks" {
  source = "./e2e"

  app_name = "${local.name}-aws-eks"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir = "aws-eks"

  install_count = 0
  aws = [
    {
      iam_role_arn = module.install_access.iam_role_arn
      regions = ["us-west-2"]
    }
  ]
}

module "aws-eks-byo-vpc" {
  source = "./e2e"

  app_name = "${local.name}-aws-byo-vpc"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir = "aws-eks-byovpc"

  install_count = 0
  aws = [
    {
      iam_role_arn = module.install_access.iam_role_arn
      regions = ["us-west-2"]
    }
  ]
}
