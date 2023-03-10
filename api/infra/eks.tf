data "aws_iam_policy_document" "api" {
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
      "s3:*Object",
    ]
    resources = ["*", ]
  }
  statement {
    effect = "Allow"
    actions = [
      "rds-db:connect",
    ]
    resources = [
      format("arn:aws:rds-db:%s:%s:dbuser:%s/%s",
        local.vars.region,
        local.target_account_id,
        module.primary.db_instance_resource_id,
        local.name
      ),
    ]
  }
}

resource "aws_iam_policy" "api_policy" {
  name   = "eks-policy-${local.name}"
  policy = data.aws_iam_policy_document.api.json
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
    custom = aws_iam_policy.api_policy.arn
  }
}
