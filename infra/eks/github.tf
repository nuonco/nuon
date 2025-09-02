data "aws_iam_policy_document" "github_actions_policy_doc" {
  statement {
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken",
      "eks:DescribeCluster",
      "eks:ListCluster",
    ]

    // NOTE(jm): we can not use `module.eks.arn` here, because that would create a circular dependency. Since we only
    // run a single cluster per account, this is effectively the same thing, regardless.
    resources = ["*", ]
  }
}

resource "aws_iam_policy" "github_actions_policy" {
  name   = "github-actions-policy-${local.workspace_trimmed}"
  policy = data.aws_iam_policy_document.github_actions_policy_doc.json
}

module "github_actions" {
  source      = "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
  version     = "5.59.0"
  create_role = true

  role_name = "github-actions-role-${local.workspace_trimmed}"

  provider_url                   = "token.actions.githubusercontent.com"
  oidc_subjects_with_wildcards   = ["repo:${local.github_organization}/${local.github_repository}:*", ]
  oidc_fully_qualified_audiences = ["sts.amazonaws.com", ]

  role_policy_arns = [
    aws_iam_policy.github_actions_policy.arn,
  ]
}
