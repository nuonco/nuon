locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  bucket_name = "nuon-dev"
  name        = "dev"
  region      = "us-west-2"

  tags = {
    service   = local.name
    terraform = local.name
  }

  networks = {
    sandbox = {
      cidr             = "10.128.0.0/16"
      public_subnets   = ["10.128.0.0/26", "10.128.0.64/26", "10.128.0.128/26"]
      private_subnets  = ["10.128.128.0/24", "10.128.129.0/24", "10.128.130.0/24"]
      database_subnets = ["10.128.131.0/24", "10.128.132.0/24"]
    }
  }
}
