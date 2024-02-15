module "eks_access" {
  source = "github.com/nuonco/sandboxes//iam-role"
  sandbox = "aws-eks"
  prefix = "workers-canary-${var.env}"
}

module "ecs_access" {
  source = "github.com/nuonco/sandboxes//iam-role"
  sandbox = "aws-ecs"
  prefix = "workers-canary-${var.env}"
}
