locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts :
    acct.name => acct.id
  }

  tf_role_arn = "arn:aws:iam::${local.accounts[var.env]}:role/terraform"

  k8s_exec = [{
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    # This requires the awscli to be installed locally where Terraform is executed
    args = ["eks", "get-token", "--cluster-name", "${var.env}-${local.vars.pool}", "--role-arn", local.tf_role_arn, ]
  }]
}

# this is the root account that the credentials have permissions for.
# use it to get list of accounts and pivot to the correct one
provider "aws" {
  region = local.vars.region
  alias  = "mgmt"
}

provider "aws" {
  region = local.vars.region

  assume_role {
    role_arn = local.tf_role_arn
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.vars.region
  alias  = "infra-shared-prod"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts["infra-shared-prod"]}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "helm" {
  kubernetes {
    host                   = nonsensitive(data.tfe_outputs.infra-eks-nuon.values.cluster_endpoint)
    cluster_ca_certificate = base64decode(data.tfe_outputs.infra-eks-nuon.values.cluster_certificate_authority_data)

    dynamic "exec" {
      for_each = local.k8s_exec
      content {
        api_version = exec.value.api_version
        command     = exec.value.command
        args        = exec.value.args
      }
    }
  }
}

provider "kubectl" {
  host                   = nonsensitive(data.tfe_outputs.infra-eks-nuon.values.cluster_endpoint)
  cluster_ca_certificate = base64decode(data.tfe_outputs.infra-eks-nuon.values.cluster_certificate_authority_data)
  apply_retry_count      = 5
  load_config_file       = false

  dynamic "exec" {
    for_each = local.k8s_exec
    content {
      api_version = exec.value.api_version
      command     = exec.value.command
      args        = exec.value.args
    }
  }
}
