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

  install_count = 0
  aws = [
    {
      iam_role_arn = module.eks_access.iam_role_arn
      regions = ["us-west-2", "us-east-1", "us-west-1", "us-east-2", "eu-west-1"]
    }
  ]
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

  install_count = 0
  aws = [
    {
      iam_role_arn = module.eks_access.iam_role_arn
      regions = ["us-west-2", "us-east-1", "us-west-1", "us-east-2", "eu-west-1"]
    }
  ]
}
