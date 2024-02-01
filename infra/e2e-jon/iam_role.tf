locals {
  eks_branch = "main"
  eks_repo = "nuonco/sandboxes"
  eks_dir = "aws-ecs"
  eks_artifact_base_url = "https://raw.githubusercontent.com/${local.repo}/${local.branch}/${local.dir}/artifacts"
}

data "http" "eks_sandbox_trust_policy" {
  url = "${local.eks_artifact_base_url}/trust.json"
}

data "http" "eks_sandbox_provision_policy" {
  url = "${local.eks_artifact_base_url}/provision.json"
}

data "http" "eks_sandbox_deprovision_policy" {
  url = "${local.eks_artifact_base_url}/deprovision.json"
}

resource "aws_iam_policy" "eks_deprovision" {
  name   = "nuon-${local.name}-install-deprovision-access-eks"
  policy = data.http.eks_sandbox_deprovision_policy.response_body
}

resource "aws_iam_policy" "eks_provision" {
  name   = "nuon-${local.name}-install-provision-access-eks"
  policy = data.http.eks_sandbox_provision_policy.response_body
}

module "install_access" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = ">= 5.1.0"

  create_role = true

  role_name = "${local.name}-customer-iam-role"

  allow_self_assume_role   = true
  custom_role_trust_policy        = data.http.eks_sandbox_trust_policy.response_body
  create_custom_role_trust_policy = true
  custom_role_policy_arns = [
    aws_iam_policy.eks_deprovision.arn,
    aws_iam_policy.eks_provision.arn
  ]
}
