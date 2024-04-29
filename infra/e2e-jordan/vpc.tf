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
      public_subnets   = ["10.128.0.0/26", "10.128.0.64/26", "10.128.0.128/26"]
      private_subnets  = ["10.128.128.0/24", "10.128.129.0/24", "10.128.130.0/24"]
      database_subnets = ["10.128.131.0/24", "10.128.132.0/24"]
    }
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

module "byovpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.5"

  name = "byoc"
  cidr = local.networks["sandbox"]["cidr"]

  azs              = data.aws_availability_zones.available.zone_ids
  private_subnets  = local.networks["sandbox"]["private_subnets"]
  public_subnets   = local.networks["sandbox"]["public_subnets"]
  database_subnets = local.networks["sandbox"]["database_subnets"]

  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  create_database_subnet_group = true
  #create_elasticache_subnet_group = true

  public_subnet_tags = {
    "visibility"             = "public",
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "visibility"                      = "private",
    "kubernetes.io/role/internal-elb" = "1"
  }
}

// module "byovpc_untagged" {
//   source  = "terraform-aws-modules/vpc/aws"
//   version = "~> 5.5"
// 
//   name = "byovpc-untagged"
//   cidr = local.networks["sandbox"]["cidr"]
// 
//   azs              = data.aws_availability_zones.available.zone_ids
//   private_subnets  = local.networks["sandbox"]["private_subnets"]
//   public_subnets   = local.networks["sandbox"]["public_subnets"]
//   database_subnets = local.networks["sandbox"]["database_subnets"]
// 
//   enable_nat_gateway   = true
//   single_nat_gateway   = true
//   enable_dns_hostnames = true
// 
//   create_database_subnet_group = true
//   #create_elasticache_subnet_group = true
// }
