locals {}

provider "nuon" {
  alias = "sandbox"
  org_id = var.sandbox_org_id
}

provider "nuon" {
  org_id = var.org_id
}

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

provider "aws" {
  alias = "tonic-test"
  region = "us-west-2"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.demo.id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}
