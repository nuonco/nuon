resource "aws_iam_policy" "install_deprovision" {
  provider = aws.canary

  name   = "nuon-${local.name}-${var.env}-install-deprovision-access"
  policy = file("policies/deprovision.json")
}

resource "aws_iam_policy" "install_provision" {
  provider = aws.canary

  name   = "nuon-${local.name}-${var.env}install-provision-access"
  policy = file("policies/provision.json")
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
  custom_role_trust_policy        = file("policies/trust.json")
  custom_role_policy_arns = [
    aws_iam_policy.install_deprovision.arn,
    aws_iam_policy.install_provision.arn
  ]
}
