locals {
  name              = "e2e-jordan"
  sandbox_name      = "aws-ecs-byovpc"
  sandboxes_repo    = "nuonco/sandboxes"
  sandbox_branch    = "main"
  artifact_base_url = "https://raw.githubusercontent.com/nuonco/sandboxes/${local.sandbox_branch}/${local.sandbox_name}/artifacts"
}

data "http" "sandbox_trust_policy" {
  url = "${local.artifact_base_url}/trust.json"
}

data "http" "sandbox_provision_policy" {
  url = "${local.artifact_base_url}/provision.json"
}

data "http" "sandbox_deprovision_policy" {
  url = "${local.artifact_base_url}/deprovision.json"
}

resource "aws_iam_policy" "deprovision" {
  name   = "nuon-${local.name}-${local.sandbox_name}-install-deprovision-access"
  policy = data.http.sandbox_deprovision_policy.response_body
}

resource "aws_iam_policy" "provision" {
  name   = "nuon-${local.name}-${local.sandbox_name}-install-provision-access"
  policy = data.http.sandbox_provision_policy.response_body
}

module "access" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = ">= 5.1.0"

  create_role = true

  role_name = "${local.name}-${local.sandbox_name}-customer-iam-role"

  allow_self_assume_role          = true
  custom_role_trust_policy        = data.http.sandbox_trust_policy.response_body
  create_custom_role_trust_policy = true
  custom_role_policy_arns = [
    aws_iam_policy.deprovision.arn,
    aws_iam_policy.provision.arn
  ]
}

module "e2e" {
  source = "../nuon"

  app_name = "${local.name}-${local.sandbox_name}"

  sandbox_repo    = local.sandboxes_repo
  sandbox_branch  = local.sandbox_branch
  sandbox_dir     = local.sandbox_name
  app_runner_type = "aws-ecs"

  east_1_count = 0
  east_2_count = 0
  west_2_count = 1

  install_role_arn = module.access.iam_role_arn
  install_inputs = [
    {
      name          = "vpc_id"
      description   = "vpc id from user"
      required      = true
      default       = ""
      value         = var.vpc_id
      interpolation = "{{.nuon.install.inputs.vpc_id}}"
    }
  ]
}
