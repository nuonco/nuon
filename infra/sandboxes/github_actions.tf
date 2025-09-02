data "aws_iam_policy_document" "github_actions_policy_doc" {
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
    ]
    resources = [module.bucket.s3_bucket_arn, ]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:*Object",
    ]
    resources = ["${module.bucket.s3_bucket_arn}/*", ]
  }
}

resource "aws_iam_policy" "github_actions_policy" {
  name   = "github-actions-policy-${local.name}"
  policy = data.aws_iam_policy_document.github_actions_policy_doc.json
}

module "github_actions" {
  source      = "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
  version     = "5.59.0"
  create_role = true

  role_name                      = "gha-${local.name}"
  role_path                      = "/github/actions/"
  provider_url                   = "token.actions.githubusercontent.com"
  oidc_subjects_with_wildcards   = ["repo:${local.github_organization}/${local.github_repository}:*", ]
  oidc_fully_qualified_audiences = ["sts.amazonaws.com", ]

  role_policy_arns = [
    aws_iam_policy.github_actions_policy.arn,
  ]
}
