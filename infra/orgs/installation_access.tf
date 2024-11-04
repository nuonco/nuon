// NOTE(jdt): create these resources in the external account
data "aws_iam_policy_document" "install_k8s" {
  statement {
    effect = "Allow"
    actions = [
      "eks:DescribeCluster",
      "eks:ListCluster",
      "sts:AssumeRole",
    ]
    resources = ["*", ]
  }
}

data "aws_iam_policy_document" "install_k8s_trust_policy_external" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole", ]

    principals {
      type = "AWS"
      identifiers = concat(
        # nuon users authenticated via sso
        # NOTE(jdt): may want to restrict this to just a readonly / support role in the future
        tolist(data.aws_iam_roles.nuon_sso_roles_external.arns),
        tolist(data.aws_iam_roles.nuon_sso_roles_workload.arns),
        [
          module.support_role.iam_role_arn
        ],
      )
    }
  }
}

resource "aws_iam_policy" "install_k8s_external" {
  provider = aws.orgs

  name   = "eks-policy-${var.env}-install"
  policy = data.aws_iam_policy_document.install_k8s.json
}

module "install_k8s_role_external" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = ">= 5.1.0"
  providers = {
    aws = aws.orgs
  }

  create_role = true

  role_name = "install-k8s-admin-${var.env}"

  create_custom_role_trust_policy = true
  custom_role_trust_policy        = data.aws_iam_policy_document.install_k8s_trust_policy_external.json
  custom_role_policy_arns         = [aws_iam_policy.install_k8s_external.arn, ]
}
