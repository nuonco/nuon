# Data source for OIDC provider from EKS cluster
data "aws_eks_cluster" "cluster" {
  name = data.tfe_outputs.infra-eks-nuon.values.cluster_name
}

data "aws_iam_openid_connect_provider" "eks" {
  url = data.aws_eks_cluster.cluster.identity[0].oidc[0].issuer
}

# Individual assume role policy for each runner scale set
data "aws_iam_policy_document" "runner_assume_role" {
  for_each = lookup(local.vars, "scale_sets", {})

  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringEquals"
      variable = "${replace(data.aws_iam_openid_connect_provider.eks.url, "https://", "")}:sub"
      values   = ["system:serviceaccount:${local.vars.runner_namespace}:${each.key}-gha-rs-kube-mode"]
    }

    condition {
      test     = "StringEquals"
      variable = "${replace(data.aws_iam_openid_connect_provider.eks.url, "https://", "")}:aud"
      values   = ["sts.amazonaws.com"]
    }

    principals {
      identifiers = [data.aws_iam_openid_connect_provider.eks.arn]
      type        = "Federated"
    }
  }
}

# IAM policy for GitHub Actions runners - only allows assuming gha- prefixed roles
data "aws_iam_policy_document" "runner_policy" {
  statement {
    sid    = "AssumeGHARoles"
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]
    resources = [
      "arn:aws:iam::${local.accounts.prod}:role/gha-*",
    ]
  }
}

# Create IAM roles for each scale set
resource "aws_iam_role" "runner_scale_set_roles" {
  for_each = lookup(local.vars, "scale_sets", {})

  name               = "${local.name}-${each.key}-runner"
  assume_role_policy = data.aws_iam_policy_document.runner_assume_role[each.key].json

  tags = merge(local.tags, {
    Name = "${local.name}-${each.key}-runner"
    ScaleSetName = each.key
  })
}

# Single IAM policy for all runner types
resource "aws_iam_policy" "runner_policy" {
  name   = "${local.name}-runner-policy"
  policy = data.aws_iam_policy_document.runner_policy.json

  tags = merge(local.tags, {
    Name = "${local.name}-runner-policy"
  })
}

# Attach the single policy to all runner roles
resource "aws_iam_role_policy_attachment" "runner_policy_attachments" {
  for_each = lookup(local.vars, "scale_sets", {})

  role       = aws_iam_role.runner_scale_set_roles[each.key].name
  policy_arn = aws_iam_policy.runner_policy.arn
}