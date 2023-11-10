provider "nuon" {}

provider "aws" {
  region = local.region
  alias  = "mgmt"
  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.region

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.demo.id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}
