provider "aws" {
  region = local.region
  alias  = "mgmt"
  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.region
  alias  = "canary"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.canary.id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.region
  alias  = "infra-shared-prod"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.infra-shared-prod.id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.region

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts[var.env].id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}
