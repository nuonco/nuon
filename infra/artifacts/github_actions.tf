data "aws_iam_policy_document" "github_actions_policy_doc" {
  // allow pushing to artifacts bucket
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
    ]
    resources = [module.bucket.s3_bucket_arn, ]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:*Object",
    ]
    resources = ["${module.bucket.s3_bucket_arn}/*", ]
  }

  // grant permissions to auth with
  statement {
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr-public:GetAuthorizationToken",
      "sts:GetServiceBearerToken",
    ]
    resources = ["*", ]
  }

  // grant permissions for public repos
  statement {
    actions = [
      "ecr-public:BatchCheckLayerAvailability",
      "ecr-public:BatchGetImage",
      "ecr-public:BatchDeleteImage",
      "ecr-public:BatchImportUpstreamImage",
      "ecr-public:CompleteLayerUpload",
      "ecr-public:DescribeImages",
      "ecr-public:DescribeRepositories",
      "ecr-public:GetDownloadUrlForLayer",
      "ecr-public:InitiateLayerUpload",
      "ecr-public:ListImages",
      "ecr-public:PutImage",
      "ecr-public:UploadLayerPart",
    ]
    resources = [
      // public cli
      module.cli.repository_arn,
      // public runner
      module.runner.repository_arn,

      // helm charts
      module.helm_demo.repository_arn,
      module.helm_temporal.repository_arn,
      module.helm_waypoint.repository_arn,

      // waypoint plugins
      module.waypoint_plugin_exp.repository_arn,
      module.waypoint_plugin_helm.repository_arn,
      module.waypoint_plugin_noop.repository_arn,
      module.waypoint_plugin_oci.repository_arn,
      module.waypoint_plugin_oci_sync.repository_arn,
      module.waypoint_plugin_terraform.repository_arn,
      module.waypoint_plugin_job.repository_arn,

      // e2e
      module.e2e.repository_arn,
    ]
  }

  // grant permissions for internal repos
  statement {
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:BatchGetImage",
      "ecr:BatchDeleteImage",
      "ecr:BatchImportUpstreamImage",
      "ecr:CompleteLayerUpload",
      "ecr:DescribeImages",
      "ecr:DescribeRepositories",
      "ecr:GetDownloadUrlForLayer",
      "ecr:InitiateLayerUpload",
      "ecr:ListImages",
      "ecr:PutImage",
      "ecr:UploadLayerPart",
    ]
    resources = [
      module.nuonctl.repository_arn,
      module.sandbox_aws_eks.repository_arn,
      module.sandbox_empty.repository_arn,
    ]
  }
}

resource "aws_iam_policy" "github_actions_policy" {
  name   = "github-actions-policy-${local.name}"
  policy = data.aws_iam_policy_document.github_actions_policy_doc.json
}

module "github_actions" {
  source      = "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
  version     = "5.59.0"
  create_role = true

  role_name                      = "gha-${local.name}"
  role_path                      = "/github/actions/"
  provider_url                   = "token.actions.githubusercontent.com"
  oidc_subjects_with_wildcards   = ["repo:${local.github.organization}/${local.github.repo}:*", ]
  oidc_fully_qualified_audiences = ["sts.amazonaws.com", ]

  role_policy_arns = [
    aws_iam_policy.github_actions_policy.arn,
  ]
}
