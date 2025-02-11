locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts :
    acct.name => acct.id
  }

  aws_settings = {
    region        = "us-west-2"
    public_region = "us-east-1"
    account_name  = "public"
  }
  # This list must align with the list of replicated provider regions
  replication_regions = [
    "ap-south-1",
    "eu-north-1",
    "eu-west-3",
    "eu-west-2",
    "eu-west-1",
    "ap-northeast-3",
    "ap-northeast-2",
    "ap-northeast-1",
    "ca-central-1",
    "sa-east-1",
    "ap-southeast-1",
    "ap-southeast-2",
    "eu-central-1",
    "us-east-1",
    "us-east-2",
    "us-west-1",
  ] # us-west-2 not included here because it is the primary region
}

provider "aws" {
  region = local.aws_settings.region

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts[local.aws_settings.account_name]}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  alias  = "public"
  region = local.aws_settings.public_region


  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

# this is the root account that the credentials have permissions for.
# use it to get list of accounts and pivot to the correct one
provider "aws" {
  region = local.aws_settings.region
  alias  = "mgmt"
}

provider "aws" {
  alias  = "infra-shared-prod"
  region = local.aws_settings.region
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.infra-shared-prod}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

# The following are all accounts used purely to replicate certain subpaths of
# the nuon-artifacts bucket into regions worldwide. The use case for this
# (initially at least) is to allow setting up cloudformation quick create links 

provider "aws" {
  alias  = "replicated-ap-south-1"
  region = "ap-south-1"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-eu-north-1"
  region = "eu-north-1"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-eu-west-3"
  region = "eu-west-3"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-eu-west-2"
  region = "eu-west-2"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-eu-west-1"
  region = "eu-west-1"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-ap-northeast-3"
  region = "ap-northeast-3"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-ap-northeast-2"
  region = "ap-northeast-2"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-ap-northeast-1"
  region = "ap-northeast-1"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-ca-central-1"
  region = "ca-central-1"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-sa-east-1"
  region = "sa-east-1"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-ap-southeast-1"
  region = "ap-southeast-1"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-ap-southeast-2"
  region = "ap-southeast-2"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-eu-central-1"
  region = "eu-central-1"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-us-east-1"
  region = "us-east-1"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-us-east-2"
  region = "us-east-2"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-us-west-1"
  region = "us-west-1"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}

provider "aws" {
  alias  = "replicated-us-west-2"
  region = "us-west-2"
  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public}:role/terraform"
  }

  default_tags {
    tags = local.replicated_tags
  }
}
