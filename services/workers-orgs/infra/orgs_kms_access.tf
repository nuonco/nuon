data "aws_iam_policy_document" "orgs_account_kms_access" {
  provider = aws.orgs

  statement {
    effect = "Allow"
    actions = [
      "kms:DescribeKey",
      "kms:DisableKey",
      "kms:CreateKey",
      "kms:CreateAlias",
      "kms:CreateGrant",
      "kms:EnableKeyRotation",
      "kms:PutKeyPolicy",
      "kms:TagResource",
      "kms:UntagResource",
    ]
    resources = ["*", ]
  }
}

data "aws_iam_policy_document" "orgs_account_kms_access_trust" {
  provider = aws.orgs

  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole", ]

    principals {
      type = "AWS"
      identifiers = [
        module.iam_eks_role.iam_role_arn,
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

resource "aws_iam_policy" "orgs_account_kms_access_policy" {
  provider = aws.orgs

  name   = "${local.name}-orgs-account-kms-access"
  policy = data.aws_iam_policy_document.orgs_account_kms_access.json
}

module "orgs_account_kms_access_role" {
  providers = {
    aws = aws.orgs
  }


  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = ">= 5.1.0"

  create_role       = true
  role_requires_mfa = false

  role_name                = "${local.name}-orgs-account-kms-access"
  custom_role_trust_policy = data.aws_iam_policy_document.orgs_account_kms_access_trust.json
  custom_role_policy_arns  = [aws_iam_policy.orgs_account_kms_access_policy.arn, ]
}
