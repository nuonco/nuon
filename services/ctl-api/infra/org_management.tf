data "aws_iam_policy_document" "orgs_account_iam_access" {
  provider = aws.orgs

  // allow management of IAM roles for an org
  statement {
    effect = "Allow"
    actions = [
      "iam:CreatePolicy",
      "iam:CreateRole",
      "iam:GetRole",
      "iam:TagRole",
      "iam:TagPolicy",
      "iam:AttachRolePolicy",
      "iam:DeletePolicy",
      "iam:DeleteRole",
      "iam:DetachRolePolicy",
    ]
    resources = ["*", ]
  }

  // allow management of ecr repos
  statement {
    effect = "Allow"
    actions = [
      "ecr:*",
    ]
    resources = ["*", ]
  }

}

data "aws_iam_policy_document" "orgs_account_iam_access_trust" {
  provider = aws.orgs

  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole", ]

    principals {
      type = "AWS"
      identifiers = [
        module.service.eks_role_arn,
        data.tfe_outputs.infra-orgs.values.iam_roles.support.arn,
      ]
    }
  }

  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole", ]

    principals {
      type = "AWS"
      identifiers = [
        data.tfe_outputs.infra-orgs.values.iam_roles.support.arn,
      ]
    }
  }
}

resource "aws_iam_policy" "orgs_account_iam_access_policy" {
  provider = aws.orgs

  name   = "${local.name}-orgs-account-iam-access"
  policy = data.aws_iam_policy_document.orgs_account_iam_access.json
}

module "management_access_role" {
  providers = {
    aws = aws.orgs
  }


  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = "5.59.0"

  create_role       = true
  role_requires_mfa = false

  role_name                       = "${local.name}-orgs-account-iam-access"
  create_custom_role_trust_policy = true
  custom_role_trust_policy        = data.aws_iam_policy_document.orgs_account_iam_access_trust.json
  custom_role_policy_arns         = [aws_iam_policy.orgs_account_iam_access_policy.arn, ]
}

