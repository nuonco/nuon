data "aws_iam_policy_document" "orgs_account_bucket_access" {
  provider = aws.orgs

  statement {
    effect = "Allow"
    actions = [
      "s3:*",
    ]
    resources = ["arn:aws:s3:::${data.tfe_outputs.infra-orgs.values.buckets.orgs.name}/", ]
  }
}

data "aws_iam_policy_document" "orgs_account_bucket_access_trust" {
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
}

resource "aws_iam_policy" "orgs_account_bucket_access_policy" {
  provider = aws.orgs

  name   = "${local.name}-orgs-bucket-access"
  policy = data.aws_iam_policy_document.orgs_account_bucket_access.json
}

module "orgs_account_bucket_access_role" {
  providers = {
    aws = aws.orgs
  }


  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = ">= 5.1.0"

  create_role       = true
  role_requires_mfa = false

  role_name                       = "${local.name}-orgs-bucket-access"
  create_custom_role_trust_policy = true
  custom_role_trust_policy        = data.aws_iam_policy_document.orgs_account_bucket_access_trust.json
  custom_role_policy_arns         = [aws_iam_policy.orgs_account_bucket_access_policy.arn, ]
}
