data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "aws_caller_identity" "current" {}
