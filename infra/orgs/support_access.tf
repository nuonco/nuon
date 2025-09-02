data "aws_iam_policy_document" "support" {
  statement {
    effect = "Allow"
    actions = [
      "eks:DescribeCluster",
      "eks:ListCluster",
      "sts:AssumeRole",
      "ecr:*",
      "s3:*",
      "kms:GenerateDataKey",
      "kms:Decrypt"
    ]
    resources = ["*", ]
  }
}

data "aws_iam_policy_document" "support_trust" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole", ]

    principals {
      type = "AWS"
      identifiers = concat(
        # nuon users authenticated via sso
        # NOTE(jdt): may want to restrict this to just a readonly / support role in the future
        tolist(data.aws_iam_roles.nuon_sso_roles_workload.arns),
      )
    }
  }
}

resource "aws_iam_policy" "support" {
  provider = aws.orgs

  name   = "eks-policy-${var.env}-install"
  policy = data.aws_iam_policy_document.support.json
}

module "support_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = "5.59.0"
  providers = {
    aws = aws.orgs
  }

  create_role = true

  role_name = "nuon-internal-support-${var.env}"

  create_custom_role_trust_policy = true
  custom_role_trust_policy        = data.aws_iam_policy_document.support_trust.json
  custom_role_policy_arns         = [aws_iam_policy.support.arn, ]
}
