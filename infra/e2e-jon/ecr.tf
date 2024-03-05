module "nuon_ecr_access" {
  source  = "nuonco/ecr-access/aws"

  role_name = "e2e-jon-aws-ecr-access"
  policy_name = "e2e-jon-aws-ecr-access"
  repository_arns = ["arn:aws:ecr:us-east-2:949309607565:repository/inl116lrng95ijmqweipezsf06"]
  providers = {
    aws = aws.tonic-test
  }
}

output "ecr_access_iam_role" {
  value = module.nuon_ecr_access.iam_role_arn
}
