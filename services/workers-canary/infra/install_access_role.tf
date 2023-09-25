locals {
  artifact_base_url = "https://nuon-artifacts.s3.us-west-2.amazonaws.com/sandbox/aws-eks"
}

data "http" "sandbox_version" {
  url = "${local.artifact_base_url}/latest.txt"
}

data "http" "sandbox_trust_policy" {
  url = "${local.artifact_base_url}/${chomp(data.http.sandbox_version.response_body)}/trust.json"
}

data "http" "sandbox_provision_policy" {
  url = "${local.artifact_base_url}/${chomp(data.http.sandbox_version.response_body)}/provision.json"
}

data "http" "sandbox_deprovision_policy" {
  url = "${local.artifact_base_url}/${chomp(data.http.sandbox_version.response_body)}/deprovision.json"
}

resource "aws_iam_policy" "install_deprovision" {
  provider = aws.canary

  name   = "nuon-${local.name}-${var.env}-install-deprovision-access"
  policy = data.http.sandbox_deprovision_policy.response_body
}

resource "aws_iam_policy" "install_provision" {
  provider = aws.canary

  name   = "nuon-${local.name}-${var.env}install-provision-access"
  policy = data.http.sandbox_provision_policy.response_body
}

module "install_access" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = ">= 5.1.0"
  providers = {
    aws = aws.canary
  }

  create_role = true

  role_name = "nuon-${local.name}-${var.env}-install-access"

  allow_self_assume_role          = true
  create_custom_role_trust_policy = true
  custom_role_trust_policy        = data.http.sandbox_trust_policy.response_body
  custom_role_policy_arns = [
    aws_iam_policy.install_deprovision.arn,
    aws_iam_policy.install_provision.arn
  ]
}
