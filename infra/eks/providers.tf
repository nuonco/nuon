data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

locals {
  tf_role_arn = "arn:aws:iam::${local.target_account_id}:role/terraform"

  k8s_exec = [{
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    # This requires the awscli to be installed locally where Terraform is executed
    args = ["eks", "get-token", "--cluster-name", module.eks.cluster_name, "--role-arn", local.tf_role_arn, ]
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

# hack. see eks.tf for more details
provider "aws" {
  region = local.vars.region
  alias  = "no_tags"
  assume_role {
    role_arn = local.tf_role_arn
  }
}

# this is the us-east-1 region used only for the aws_ecrpublic_authorization_token
provider "aws" {
  region = "us-east-1"
  alias  = "virginia"
}

provider "twingate" {}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

  dynamic "exec" {
    for_each = local.k8s_exec
    content {
      api_version = exec.value.api_version
      command     = exec.value.command
      args        = exec.value.args
    }
  }
}

provider "helm" {
  debug = true
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

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
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
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

provider "random" {}

