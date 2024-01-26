# Create a VPC for byovpc sandboxes to use.

locals {
  networks = {
    # each network block is configured by taking a /16 and dividing by 2
    # this leaves 2 /17's
    # public subnets are taken from the first /17
    # private_subnets are taken from the second half
    sandbox = {
      cidr = "10.128.0.0/16"
      # public subnets do not need to be as big so we evenly break down a /24
      # 10.128.0.192/26 finishes  the /24 segementation
      # you can see the math for this /16 here https://www.davidc.net/sites/default/subnets/subnets.html?network=10.128.0.0&mask=16&division=23.ff3100
      public_subnets  = ["10.128.0.0/26", "10.128.0.64/26", "10.128.0.128/26"]
      private_subnets = ["10.128.128.0/24", "10.128.129.0/24", "10.128.130.0/24"]
    }
  }
}

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "byovpc"
  cidr = local.networks["sandbox"]["cidr"]

  azs             = [for az in ["a", "b", "c"] : "${local.region}${az}"]
  private_subnets = local.networks["sandbox"]["private_subnets"]
  public_subnets  = local.networks["sandbox"]["public_subnets"]

  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true
  #TODO(jm): these might be breaking installs
  #create_database_subnet_group = true
  #create_elasticache_subnet_group = true

  default_security_group_egress = [
    {
      "protocol"  = "-1"
      from_port   = 0
      to_port     = 0
      cidr_blocks = "0.0.0.0/0"
    }
  ]

  default_security_group_ingress = [
    {
      "protocol"  = "-1"
      from_port   = 0
      to_port     = 0
      cidr_blocks = "0.0.0.0/0"
    }
  ]

  public_subnet_tags = {
    "kubernetes.io/role/elb" = 1
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
  }
}
