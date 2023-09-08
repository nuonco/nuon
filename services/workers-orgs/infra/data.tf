data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

data "aws_ecr_repository" "ecr_repository" {
  provider = aws.infra-shared-prod
  name     = "mono/${local.name}"
}
