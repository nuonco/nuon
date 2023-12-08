module "aws-eks" {
  source = "./e2e"

  app_name = "${local.name}-aws-eks"

  sandbox_repo   = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir    = "aws-byo-vpc"

  east_1_count = 0
  east_2_count = 0
  west_2_count = 0

  install_role_arn = module.install_access.iam_role_arn
}
