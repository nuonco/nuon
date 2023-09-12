data "aws_iam_policy_document" "extra_auth_map" {
  for_each = { for add in local.vars.auth_map_additions : add.name => add }
  statement {
    effect = "Allow"
    actions = [
      "eks:DescribeCluster",
      "eks:ListClusters",
    ]
    resources = [module.eks.cluster_arn, ]
  }
}

data "aws_iam_policy_document" "extra_auth_map_trust_policy" {
  for_each = { for add in local.vars.auth_map_additions : add.name => add }
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole", ]

    principals {
      type = "AWS"
      identifiers = concat(each.value.trust, [
        format("arn:aws:iam::%s:role/eks/%s", local.accounts[each.value.account].id, each.key)
      ])
    }
  }
}

resource "aws_iam_policy" "extra_auth_map" {
  for_each = { for add in local.vars.auth_map_additions : add.name => add }

  name   = "eks-policy-extra-auth-entry-${each.key}-${local.workspace_trimmed}"
  policy = data.aws_iam_policy_document.extra_auth_map[each.key].json
}

module "extra_auth_map" {
  for_each = { for add in local.vars.auth_map_additions : add.name => add }
  source   = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version  = ">= 5.1.0"

  create_role = true

  role_name = "extra-auth-${each.key}-${local.workspace_trimmed}"

  create_custom_role_trust_policy = true
  custom_role_trust_policy        = data.aws_iam_policy_document.extra_auth_map_trust_policy[each.key].json
  custom_role_policy_arns         = [aws_iam_policy.extra_auth_map[each.key].arn, ]
}
