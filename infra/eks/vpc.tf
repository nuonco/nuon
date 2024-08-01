locals {
  networks = {
    # each network block is configured by taking a /16 and dividing by 2
    # this leaves 2 /17's
    # public subnets are taken from the first /17
    # private_subnets are taken from the second half
    ci-nuon = {
      cidr            = "10.129.0.0/16"
      public_subnets  = ["10.129.0.0/26", "10.129.0.64/26", "10.129.0.128/26"]
      private_subnets = ["10.129.128.0/24", "10.129.129.0/24", "10.129.130.0/24"]
    }

    stage-nuon = {
      cidr            = "10.131.0.0/16"
      public_subnets  = ["10.131.0.0/26", "10.131.0.64/26", "10.131.0.128/26"]
      private_subnets = ["10.131.128.0/24", "10.131.129.0/24", "10.131.130.0/24"]
    }

    prod-nuon = {
      cidr            = "10.132.0.0/16"
      public_subnets  = ["10.132.0.0/26", "10.132.0.64/26", "10.132.0.128/26"]
      private_subnets = ["10.132.128.0/24", "10.132.129.0/24", "10.132.130.0/24"]
    }

    orgs-stage-main = {
      cidr            = "10.133.0.0/16"
      public_subnets  = ["10.133.0.0/26", "10.133.0.64/26", "10.133.0.128/26"]
      private_subnets = ["10.133.128.0/24", "10.133.129.0/24", "10.133.130.0/24"]
    }

    orgs-prod-main = {
      cidr            = "10.134.0.0/16"
      public_subnets  = ["10.134.0.0/26", "10.134.0.64/26", "10.134.0.128/26"]
      private_subnets = ["10.134.128.0/24", "10.134.129.0/24", "10.134.130.0/24"]
    }

    runners-stage-main = {
      cidr            = "10.135.0.0/16"
      public_subnets  = ["10.135.0.0/26", "10.135.0.64/26", "10.135.0.128/26"]
      private_subnets = ["10.135.128.0/24", "10.135.129.0/24", "10.135.130.0/24"]
    }

    runners-prod-main = {
      cidr            = "10.136.0.0/16"
      public_subnets  = ["10.136.0.0/26", "10.136.0.64/26", "10.136.0.128/26"]
      private_subnets = ["10.136.128.0/24", "10.136.129.0/24", "10.136.130.0/24"]
    }

    # to create a new environment with a new network add a segement exactly the same as the one above but bump the /16
    # octet up by one
  }
}


module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = local.workspace_trimmed
  cidr = local.networks[local.workspace_trimmed]["cidr"]

  azs             = [for az in ["a", "b", "c"] : "${local.region}${az}"]
  private_subnets = local.networks[local.workspace_trimmed]["private_subnets"]
  public_subnets  = local.networks[local.workspace_trimmed]["public_subnets"]

  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  public_subnet_tags = {
    "kubernetes.io/cluster/${local.workspace_trimmed}" = "shared"
    "kubernetes.io/role/elb"                           = 1
  }

  private_subnet_tags = {
    "kubernetes.io/cluster/${local.workspace_trimmed}" = "shared"
    "kubernetes.io/role/internal-elb"                  = 1
    # Tags subnets for Karpenter auto-discovery
    (local.karpenter.discovery_key) = local.karpenter.discovery_value
  }
}

module "endpoints" {
  source  = "terraform-aws-modules/vpc/aws//modules/vpc-endpoints"
  version = "~> 5.0"

  vpc_id = module.vpc.vpc_id

  endpoints = {
    s3 = {
      service      = "s3"
      service_type = "Gateway"
      route_table_ids = flatten([
        module.vpc.intra_route_table_ids,
        module.vpc.private_route_table_ids,
        module.vpc.public_route_table_ids,
      ])
      tags = { Name = "s3-vpc-endpoint" }
    },
  }
}
