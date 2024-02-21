module "eks_access" {
  source = "github.com/nuonco/sandboxes//iam-role"
  sandbox = "aws-eks"
  prefix = "workers-canary-${var.env}"

  providers = {
    aws = aws.canary
  }
}

module "ecs_access" {
  source = "github.com/nuonco/sandboxes//iam-role"
  sandbox = "aws-ecs"
  prefix = "workers-canary-${var.env}"

  providers = {
    aws = aws.canary
  }
}
