module "aws-ecs" {
  source = "./e2e"

  app_name = "${local.name}-aws-ecs"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = "main"
  sandbox_dir = "aws-ecs"
  app_runner_type = "aws-ecs"

  east_1_count = 0
  eu_west_2_count = 1
  east_2_count = 0
  west_2_count = 0

  install_role_arn = module.ecs_access.iam_role_arn
}
