data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "aws_iam_roles" "nuon_sso_roles_workload" {
  provider    = aws.stage
  name_regex  = "AWSReservedSSO_Nuon.*"
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}
