// customer install access roles for testing
module "ecs_access" {
  source = "nuonco/install-access/aws"
  sandbox = "aws-ecs"
  prefix = "dev-ecs"
  enable_support_access = true

  providers = {
    aws = aws.demo
  }
}

module "eks_access" {
  source = "nuonco/install-access/aws"
  sandbox = "aws-ecs"
  prefix = "dev-eks"
  enable_support_access = true

  providers = {
    aws = aws.demo
  }
}

module "eks_byovpc_access" {
  source = "nuonco/install-access/aws"
  sandbox = "aws-eks-byovpc"
  prefix = "dev-eks-byovpc"
  enable_support_access = true

  providers = {
    aws = aws.demo
  }
}

module "ecs_byovpc_access" {
  source = "nuonco/install-access/aws"
  sandbox = "aws-ecs-byovpc"
  prefix = "dev-ecs-byovpc"
  enable_support_access = true
  
  providers = {
    aws = aws.demo
  }
}
