module "common-fate" {
  source = "nuonco/install-access/aws"
  sandbox = "aws-ecs"
  prefix = "common-fate"
  enable_support_access = true
}

module "warpstream" {
  source = "nuonco/install-access/aws"
  sandbox = "aws-ecs"
  prefix = "warpstream"
  enable_support_access = true
}
