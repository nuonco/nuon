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

  install_role_arn = module.ecs_access.iam_role_arn
  install_inputs = [
    {
      name = "vpc_id"
      description = "vpc id from user"
      required = true
      default = ""
      value = "vpc-0e54b0ce97f3d67dc"
      interpolation = "{{.nuon.install.inputs.vpc_id}}"
    }
  ]
}
