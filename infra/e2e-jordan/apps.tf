module "aws-ecs" {
  source = "./aws-ecs"
}

module "aws-ecs-byovpc" {
  source = "./aws-ecs-byovpc"
  vpc_id = module.byovpc.vpc_id
}
