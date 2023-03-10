data "aws_iam_policy_document" "service_policy" {
  statement {
    effect = "Allow"
    # access to read/write chart bucket
    actions = [
      "s3:ListBucket",
      "s3:*Object",
    ]
    resources = ["*", ]
  }
  statement {
    effect = "Allow"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey"
    ]
    resources = ["*", ]
  }
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole"
    ]
    resources = ["*", ]
  }

}

resource "aws_iam_policy" "service" {
  name   = "eks-policy-${local.name}"
  policy = data.aws_iam_policy_document.service_policy.json
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
    custom = aws_iam_policy.service.arn
  }
}
