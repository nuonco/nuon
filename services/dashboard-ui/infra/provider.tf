provider "aws" {
  region = local.vars.region
  alias  = "mgmt"
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
