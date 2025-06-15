locals {
  github = {
    uri = "token.actions.githubusercontent.com"
  }
}

data "aws_iam_openid_connect_provider" "github" {
  url = "https://${local.github.uri}"
}

data "aws_iam_policy_document" "github_actions_assume_role" {
  # allow GH actions that have assumed a workflow specific role to assume this role
  # for e.g. deploying to k8s
  statement {
    actions = [
      "sts:AssumeRole",
      "sts:AssumeRoleWithWebIdentity",
      "sts:TagSession",
    ]

    # this is pretty broad and typically not a great idea,
    # unfortunately, it's not really tenable to list all of the
    # principals here as wildcards aren't supported for iam/sts principals
    principals {
      type        = "AWS"
      identifiers = ["*", ]
    }

    # limit the principals that can assume this role to those coming in via
    # our Github OIDC setup in _this_ account
    condition {
      test     = "StringEquals"
      variable = "aws:FederatedProvider"
      values   = [data.aws_iam_openid_connect_provider.github.arn, ]
    }
  }
}

data "aws_iam_policy_document" "github_actions_policy" {
  statement {
    actions   = ["eks:DescribeCluster", ]
    resources = [module.eks.cluster_arn, ]
  }

  statement {
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:BatchGetImage",
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
    resources = ["*", ]
  }
}

resource "aws_iam_role" "github_actions" {
  name               = "github-actions-role-${local.workspace_trimmed}"
  assume_role_policy = data.aws_iam_policy_document.github_actions_assume_role.json
}

resource "aws_iam_policy" "github_actions" {
  name = "github-actions-${local.workspace_trimmed}"

  policy = data.aws_iam_policy_document.github_actions_policy.json
}

resource "aws_iam_role_policy_attachment" "github_actions" {
  role       = aws_iam_role.github_actions.name
  policy_arn = aws_iam_policy.github_actions.arn
}
