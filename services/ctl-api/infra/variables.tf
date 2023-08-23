data "utils_deep_merge_yaml" "vars" {
  input = [
    file("vars/defaults.yaml"),
    file("vars/${var.env}.yaml"),
  ]
}

locals {
  name = "ctl-api"
  vars = yamldecode(data.utils_deep_merge_yaml.vars.output)
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }
  region = "us-west-2"
  tags = {
    service   = local.name
    terraform = "${local.name}-${var.env}"
  }
}

variable "env" {
  type        = string
  description = "env"
}
