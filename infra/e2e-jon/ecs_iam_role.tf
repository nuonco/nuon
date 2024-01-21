locals {
  branch = "main"
  ecs_artifact_base_url = "https://raw.githubusercontent.com/nuonco/sandboxes/${local.branch}/aws-ecs-byovpc/artifacts"
}

data "http" "ecs_sandbox_trust_policy" {
  url = "${local.ecs_artifact_base_url}/trust.json"
}

data "http" "ecs_sandbox_provision_policy" {
  url = "${local.ecs_artifact_base_url}/provision.json"
}

data "http" "ecs_sandbox_deprovision_policy" {
  url = "${local.ecs_artifact_base_url}/deprovision.json"
}

resource "aws_iam_policy" "ecs_deprovision" {
  name   = "nuon-${local.name}-install-deprovision-access-ecs"
  policy = data.http.ecs_sandbox_deprovision_policy.response_body
}

resource "aws_iam_policy" "ecs_provision" {
  name   = "nuon-${local.name}-install-provision-access-ecs"
  policy = data.http.ecs_sandbox_provision_policy.response_body
}

module "ecs_access" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = ">= 5.1.0"

  create_role = true

  role_name = "${local.name}-customer-iam-role-ecs"

  allow_self_assume_role   = true
  custom_role_trust_policy        = data.http.ecs_sandbox_trust_policy.response_body
  create_custom_role_trust_policy = true
  custom_role_policy_arns = [
    aws_iam_policy.ecs_deprovision.arn,
    aws_iam_policy.ecs_provision.arn
  ]
}
