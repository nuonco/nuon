module "aws-ecs" {
  source = "./e2e"

  app_name = "${local.name}-aws-ecs"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = "jm/aws-ecs-byo-vpc"
  sandbox_dir = "aws-ecs-byovpc"
  app_runner_type = "aws-ecs"

  east_1_count = 1
  east_2_count = 0
  west_2_count = 0

  install_role_arn = module.install_access.iam_role_arn
}

