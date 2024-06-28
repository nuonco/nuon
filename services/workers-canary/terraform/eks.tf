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
  install_count = var.install_count
  aws = [
    {
      iam_role_arn = var.aws_eks_iam_role_arn
      regions = random_shuffle.aws_eks_regions.result
    }
  ]
}
