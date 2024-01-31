module "aws-ecs" {
  source = "./aws-ecs"
}

module "aws-ecs-byovpc" {
  source = "./aws-ecs-byovpc"
  vpc_id = "vpc-0ea815210f55e3c7c"
}
