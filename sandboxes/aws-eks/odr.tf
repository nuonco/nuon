data "aws_iam_policy_document" "odr" {
  statement {
    effect    = "Allow"
    actions   = ["*"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "odr" {
  name   = "odr-${local.vars.id}"
  policy = data.aws_iam_policy_document.odr.json
}

module "odr_iam_role" {
  # NOTE: the iam role requires the cluster be created, but you can not reference the cluster module in the for_each
  # loop that the eks module uses to iterate over cluster_service_accounts
  depends_on = [module.eks, aws_iam_policy.odr]

  source      = "terraform-aws-modules/iam/aws//modules/iam-eks-role"
  version     = ">= 5.1.0"
  create_role = true

  role_name = "odr-${local.vars.id}"
  role_path = "/nuon/"

  cluster_service_accounts = {
    (local.vars.id) = ["${var.waypoint_odr_namespace}:${var.waypoint_odr_service_account_name}"]
  }

  role_policy_arns = {
    custom = aws_iam_policy.odr.arn
  }
}
