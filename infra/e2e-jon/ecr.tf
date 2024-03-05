data "aws_iam_policy_document" "nuon_ecr_access" {
  statement {
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken",
    ]
    resources = ["*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "ecr:BatchGetImage",
      "ecr:ListImages",
      "ecr:ListTagsForResource",
      "ecr:GetDownloadUrlForLayer",
      "ecr:DescribeImageReplicationStatus",
      "ecr:DescribeImageScanFindings",
      "ecr:DescribeImages",
      "ecr:DescribePullThroughCacheRules",
      "ecr:DescribeRegistry",
      "ecr:DescribeRepositories",
    ]
    resources = ["arn:aws:ecr:us-east-2:949309607565:repository/inl116lrng95ijmqweipezsf06"]
  }
}

resource "aws_iam_policy" "nuon_ecr_access" {
  provider = aws.tonic-test
  name   = "aws-ecr-access"
  policy = data.aws_iam_policy_document.nuon_ecr_access.json
}

module "nuon_ecr_access" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = ">= 5.1.0"

  create_role = true

  role_name = "nuon-ecr-access"

  allow_self_assume_role   = false
  custom_role_trust_policy = file("trust.json")
  create_custom_role_trust_policy = true
  custom_role_policy_arns = [
    aws_iam_policy.nuon_ecr_access.arn
  ]

  providers = {
    aws = aws.tonic-test
  }
}

output "ecr_access_iam_role" {
  value = module.nuon_ecr_access.iam_role_arn
}
