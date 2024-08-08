data "aws_iam_policy_document" "additional_permissions" {
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
      "s3:*Object",
      "sts:AssumeRole",
    ]
    resources = ["*", ]
  }
}

resource "aws_iam_policy" "additional_permissions" {
  name   = "eks-policy-${local.name}-additional"
  policy = data.aws_iam_policy_document.additional_permissions.json
}
