data "aws_iam_policy_document" "github_actions_policy_doc" {
  statement {
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken",
      "eks:DescribeCluster",
      "eks:ListCluster",
    ]
    resources = ["*", ]
  }

  statement {
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:BatchGetImage",
      "ecr:BatchDeleteImage",
      "ecr:BatchImportUpstreamImage",
      "ecr:CompleteLayerUpload",
      "ecr:DescribeImages",
      "ecr:DescribeRepositories",
      "ecr:GetDownloadUrlForLayer",
      "ecr:InitiateLayerUpload",
      "ecr:ListImages",
      "ecr:PutImage",
      "ecr:UploadLayerPart",
    ]
    resources = [data.aws_ecr_repository.ecr_repository.arn]
  }
}

resource "aws_iam_policy" "github_actions_policy" {
  name   = "github-actions-policy-${var.name}"
  policy = data.aws_iam_policy_document.github_actions_policy_doc.json
}

module "github_actions" {
  source      = "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
  version     = "5.59.0"
  create_role = true

  role_name = "gha-${var.name}"
  role_path = "/github/actions/"

  provider_url                   = "token.actions.githubusercontent.com"
  oidc_subjects_with_wildcards   = ["repo:${local.github_organization}/${local.github_repository}:*", ]
  oidc_fully_qualified_audiences = ["sts.amazonaws.com", ]

  role_policy_arns = [
    aws_iam_policy.github_actions_policy.arn,
  ]
}
