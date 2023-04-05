data "aws_iam_policy_document" "service" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole"
    ]
    resources = ["*", ]
  }
}

resource "aws_iam_policy" "service" {
  name   = "eks-policy-${var.name}"
  policy = data.aws_iam_policy_document.service.json
}

module "iam_eks_role" {
  source      = "terraform-aws-modules/iam/aws//modules/iam-eks-role"
  version     = ">= 5.1.0"
  create_role = true

  role_name = "eks-${var.name}"
  role_path = "/eks/"

  cluster_service_accounts = {
    (local.vars.cluster_name) = ["default:${var.name}", ]
  }

  role_policy_arns = {
    custom = aws_iam_policy.service.arn
  }
}
