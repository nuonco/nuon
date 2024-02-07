module "ecs_access" {
  source = "github.com/nuonco/sandboxes//iam-role"
  sandbox = "aws-ecs"
  prefix = "e2e-jon"
}
