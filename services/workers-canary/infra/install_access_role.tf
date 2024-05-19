module "eks_access" {
  source = "nuonco/install-access/aws"

  sandbox = "aws-eks"
  prefix = "workers-canary-${var.env}"

  providers = {
    aws = aws.canary
  }
}

module "ecs_access" {
  source = "nuonco/install-access/aws"

  sandbox = "aws-ecs"
  prefix = "workers-canary-${var.env}"

  providers = {
    aws = aws.canary
  }
}
