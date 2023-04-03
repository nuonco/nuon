data "aws_iam_policy_document" "api_gateway" {
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
    ]
    resources = ["*", ]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:*Object",
    ]
    resources = ["*", ]
  }
}

resource "aws_iam_policy" "api_gateway_policy" {
  name   = "eks-policy-${local.name}"
  policy = data.aws_iam_policy_document.api_gateway.json
}

module "iam_eks_role" {
  source      = "terraform-aws-modules/iam/aws//modules/iam-eks-role"
  version     = ">= 5.1.0"
  create_role = true

  role_name = "eks-${local.name}"
  role_path = "/eks/"

  cluster_service_accounts = {
    (local.vars.cluster_name) = ["default:${local.name}", ]
  }

  role_policy_arns = {
    custom = aws_iam_policy.api_gateway_policy.arn
  }
}
