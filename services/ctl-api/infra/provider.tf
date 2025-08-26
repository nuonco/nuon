provider "aws" {
  region = local.vars.region
  alias  = "mgmt"
  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.vars.region
  alias  = "canary"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.canary.id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.vars.region
  alias  = "infra-shared-prod"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.infra-shared-prod.id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.vars.region

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts[var.env].id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  alias  = "public"
  region = "us-east-1"


  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public.id}:role/terraform"
  }

  default_tags {
    tags = {
      public = "true"
    }
  }
}

provider "aws" {
  region = local.vars.region
  alias  = "orgs"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts["orgs-${var.env}"].id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.vars.region
  alias  = "default"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts[var.env].id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}
