data "aws_iam_policy_document" "public_dns_access" {
  provider = aws.public

  statement {
    effect = "Allow"
    // TODO: lock these down once we know what permissions are needed
    actions = [
      "route53:*",
    ]
    // TODO: lock these down once we know what permissions are needed
    resources = ["*", ]
  }
}

resource "aws_iam_policy" "public_dns_access_policy" {
  provider = aws.public

  name   = "${local.name}-${var.env}-public-account-dns-access"
  policy = data.aws_iam_policy_document.public_dns_access.json
}

data "aws_iam_policy_document" "public_dns_access_trust" {
  provider = aws.public

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
}

module "public_dns_access_role" {
  providers = {
    aws = aws.public
  }

  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = "5.59.0"

  create_role       = true
  role_requires_mfa = false

  role_name                       = "${local.name}-${var.env}-public-dns-access"
  create_custom_role_trust_policy = true
  custom_role_trust_policy        = data.aws_iam_policy_document.public_dns_access_trust.json
  custom_role_policy_arns         = [aws_iam_policy.public_dns_access_policy.arn, ]
}

