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

# Enhanced assume role policy that includes both OIDC and self-hosted runners
data "aws_iam_policy_document" "github_actions_assume_role_policy" {
  # Preserve the original OIDC trust policy
  statement {
    effect  = "Allow"
    actions = [
      "sts:AssumeRoleWithWebIdentity",
      "sts:TagSession"
    ]
    
    principals {
      type        = "Federated"
      identifiers = ["arn:aws:iam::${local.accounts[var.env].id}:oidc-provider/token.actions.githubusercontent.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${local.github_organization}/${local.github_repository}:*"]
    }
  }

  # Add trust policy for self-hosted runners
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${local.accounts["infra-shared-ci"].id}:root"]
    }
    condition {
      test     = "StringLike"
      variable = "aws:PrincipalArn"
      values   = ["arn:aws:iam::${local.accounts["infra-shared-ci"].id}:role/*-runner"]
    }
  }
}

# Create the role ourselves with the enhanced assume role policy
resource "aws_iam_role" "github_actions" {
  name               = "gha-${var.name}"
  path               = "/github/actions/"
  assume_role_policy = data.aws_iam_policy_document.github_actions_assume_role_policy.json
}

# Use the module but reference our manually created role
module "github_actions" {
  source      = "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
  version     = "5.59.0"
  create_role = false  # Don't create the role, we already have it

  role_name = aws_iam_role.github_actions.name  # Reference our custom role

  provider_url                   = "token.actions.githubusercontent.com"
  oidc_subjects_with_wildcards   = ["repo:${local.github_organization}/${local.github_repository}:*"]
  oidc_fully_qualified_audiences = ["sts.amazonaws.com"]

  role_policy_arns = [
    aws_iam_policy.github_actions_policy.arn,
  ]
}

# Ensure the policy is attached to our custom role
resource "aws_iam_role_policy_attachment" "github_actions_policy" {
  role       = aws_iam_role.github_actions.name
  policy_arn = aws_iam_policy.github_actions_policy.arn
}
