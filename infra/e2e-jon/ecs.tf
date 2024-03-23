module "aws-ecs" {
  source = "./e2e"

  app_name = "${local.name}-aws-ecs"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = "main"
  sandbox_dir = "aws-ecs"
  app_runner_type = "aws-ecs"

  install_count = 0
  aws = [
    {
      iam_role_arn = module.ecs_access.iam_role_arn
      regions = ["us-west-2", "us-east-1", "us-west-1", "us-east-2", "eu-west-1"]
    }
  ]
}
