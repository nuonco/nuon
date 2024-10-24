module "nuon-vpc" {

  source  = "nuonco/vpc/aws"
  version = "~> 1.1.0"

  name = "byovpc"
  cidr = local.networks["sandbox"]["cidr"]

  private_subnets  = local.networks["sandbox"]["private_subnets"]
  public_subnets   = local.networks["sandbox"]["public_subnets"]
  database_subnets = local.networks["sandbox"]["database_subnets"]
}
