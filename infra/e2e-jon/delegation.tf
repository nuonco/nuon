module "delegation" {
  source = "nuonco/install-access-delegation/aws"
  name = "e2e-jon-delegation-access-test"
  enable_support_access = true
}

module "delegation_access" {
  source = "nuonco/install-access/aws"
  sandbox = "aws-ecs"
  prefix = "e2e-jon-delegation"
  enable_support_access = false
  delegation_role_arn = module.delegation.iam_role_arn
}
