module "dev_ecr_access" {
  source = "nuonco/ecr-access/aws"

  repository_arns = [module.dev_ecr.repository_arn]
  policy_name = "dev-nuon-ecr-access"
  role_name = "dev-nuon-ecr-access"

  providers = {
    aws = aws.demo
  }
}
