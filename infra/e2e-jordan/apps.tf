module "aws_ecs" {
  source = "./aws-ecs"
}

module "aws_ecs_byovpc" {
  source = "./aws-ecs-byovpc"
  vpc_id = module.byovpc.vpc_id
}

module "aws_eks" {
  source = "./aws-eks"
}

module "aws_eks_byovpc" {
  source = "./aws-eks-byovpc"
  vpc_id = module.byovpc.vpc_id
}
