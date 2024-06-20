module "ecs_access" {
  source = "nuonco/install-access/aws"
  sandbox = "aws-ecs"
  prefix = "e2e-jon"
  enable_support_access = true
}
