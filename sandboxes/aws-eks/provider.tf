locals {
  k8s_exec = [{
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws-iam-authenticator"
    # This requires the aws iam authenticator to be installed locally where Terraform is executed
    args = ["token", "-i", module.eks.cluster_name, "-r", var.assume_role_arn]
  }]
}

provider "aws" {
  region = local.vars.region

  assume_role {
    role_arn = var.assume_role_arn
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
    role_arn = var.assume_role_arn
  }
}

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
  apply_retry_count      = 5
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
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
