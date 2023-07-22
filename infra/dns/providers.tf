provider "aws" {
  region = var.aws_region
}

provider "aws" {
  region = var.aws_region
  alias  = "mgmt"
  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.regions.prod
  alias  = "prod"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.prod.id}:role/terraform"
  }

  default_tags {
    tags = merge({ env : "prod" }, local.tags)
  }
}

provider "aws" {
  region = local.regions.stage
  alias  = "stage"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.stage.id}:role/terraform"
  }

  default_tags {
    tags = merge({ env : "stage" }, local.tags)
  }
}

provider "aws" {
  region = local.regions.horizon
  alias  = "public"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.public.id}:role/terraform"
  }
}
