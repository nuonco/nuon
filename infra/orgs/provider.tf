provider "aws" {
  region = local.region
  alias  = "mgmt"
  default_tags {
    tags = local.tags
  }
}

// NOTE(jdt): this is for creating external facing roles in the `external` account
provider "aws" {
  region = local.region
  alias  = "external"
  default_tags {
    tags = local.tags
  }

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts["external"].id}:role/terraform"
  }
}

// NOTE(jdt): this is for fetching sso roles in the workload account
provider "aws" {
  region = local.region
  alias  = "workload"
  default_tags {
    tags = local.tags
  }

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts[var.env].id}:role/terraform"
  }
}

provider "aws" {
  region = local.region

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts[local.target_account].id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.region
  alias  = "orgs"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts["orgs-${var.env}"].id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.region
  alias  = "public"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts["public"].id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}
