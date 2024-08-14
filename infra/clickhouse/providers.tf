locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts :
    acct.name => acct.id
  }

  tf_role_arn = "arn:aws:iam::${local.accounts[var.env]}:role/terraform"

  k8s_exec = [{
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    # This requires the awscli to be installed locally where Terraform is executed
    args = ["eks", "get-token", "--cluster-name", "${var.env}-${local.vars.pool}", "--role-arn", local.tf_role_arn, ]
  }]
}

# this is the root account that the credentials have permissions for.
# use it to get list of accounts and pivot to the correct one
provider "aws" {
  region = local.vars.region
  alias  = "mgmt"
}

provider "aws" {
  region = local.vars.region

  assume_role {
    role_arn = local.tf_role_arn
  }

  default_tags {
    tags = local.tags
  }
}
