data "aws_iam_policy_document" "runner_dev" {
  statement {
    effect    = "Allow"
    actions   = ["*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "runner_dev_trust" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole", ]

    principals {
      type = "AWS"
      identifiers = concat(
        # nuon users authenticated via sso
        tolist(data.aws_iam_roles.nuon_sso_roles_workload.arns),
      )
    }
  }
}

resource "aws_iam_policy" "runner_dev" {
  provider = aws.demo

  name   = "runner-dev"
  policy = data.aws_iam_policy_document.runner_dev.json
}

module "runner_dev" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = "5.59.0"
  providers = {
    aws = aws.demo
  }

  create_role = true

  role_name = "nuon-runner-dev"

  create_custom_role_trust_policy = true
  custom_role_trust_policy        = data.aws_iam_policy_document.runner_dev_trust.json
  custom_role_policy_arns         = [aws_iam_policy.runner_dev.arn, ]
}
