data "aws_iam_policy_document" "service" {
  // NOTE(jm): the following policies are essentially a giant hack so I can fix installs.
  //
  // TLDR -- we've been migrating all services to use this module, and are hitting some limits where a few services need
  // custom permissions, beyond the standard. Ideally, we could just have the services define their own policy to pass
  // in, and then everything else here would just work. For now, we can just hard code the permissions, though.
  statement {
    effect = "Allow"
    actions = [
      "s3:CreateBucket",
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
  name   = "eks-policy-${var.name}"
  policy = data.aws_iam_policy_document.service.json
}

module "iam_eks_role" {
  source      = "terraform-aws-modules/iam/aws//modules/iam-eks-role"
  version     = "5.59.0"
  create_role = true

  role_name = "eks-${var.name}"
  role_path = "/eks/"

  cluster_service_accounts = {
    (local.vars.cluster_name) = ["${var.namespace}:${var.name}", ]
  }

  role_policy_arns = merge(
    {
      custom = aws_iam_policy.service.arn
    },
    { for i, policy in var.additional_iam_policies : tostring(i) => policy }
  )
}
