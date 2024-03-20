module "service" {
  source = "../../../infra/modules/service"

  name = "workers-canary"
  env  = var.env

  additional_iam_policies = [
    aws_iam_policy.s3_permissions.arn
  ]
}

data "aws_iam_policy_document" "s3_permissions" {
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
      "s3:*Object",
    ]
    resources = [module.canary_bucket.s3_bucket_arn]
  }
}

resource "aws_iam_policy" "s3_permissions" {
  name   = "${local.name}-${var.env}-s3"
  policy = data.aws_iam_policy_document.s3_permissions.json
}
