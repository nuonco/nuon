data "aws_iam_policy_document" "odr_sync_artifacts" {
  statement {
    effect = "Allow"
    actions = [
      "ecr:*",
    ]
    resources = [module.ecr.repository_arn]
  }

  statement {
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken",
    ]
    resources = ["*", ]
  }
}

resource "aws_iam_policy" "odr_sync_artifacts" {
  name   = "odr-sync-artifacts-${local.vars.id}"
  policy = data.aws_iam_policy_document.odr_sync_artifacts.json
}

module "odr_sync_artifacts_iam_role" {
  # NOTE: the iam role requires the cluster be created, but you can not reference the cluster module in the for_each
  # loop that the eks module uses to iterate over cluster_service_accounts
  depends_on = [module.eks]

  source      = "terraform-aws-modules/iam/aws//modules/iam-eks-role"
  version     = ">= 5.1.0"
  create_role = true

  role_name = "odr-sync-artifacts-${local.vars.id}"
  role_path = "/nuon/"

  cluster_service_accounts = {
    (local.vars.id) = ["${var.waypoint_odr_namespace}:${var.waypoint_odr_service_account_name}"]
  }

  role_policy_arns = {
    custom = aws_iam_policy.odr_sync_artifacts.arn
  }
}
