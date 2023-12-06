# Temporary test of the byovpc sandbox. Wanted to get his merged in for visibility.
# Will refactor to use the e2e module once this is working.

locals {
  cluster_name = "byovpc_install"
}

resource "nuon_app" "byovpc" {
  name = "jordan_byovpc_app"
}

resource "nuon_app_input" "byovpc" {
  app_id = nuon_app.byovpc.id

  input {
    name        = "vpc_id"
    description = "The VPC to deploy the app to."
    default     = ""
    required    = true
  }

  input {
    name        = "eks_version"
    description = "The Kubernetes version to use for the EKS cluster."
    default     = ""
    required    = true
  }

  input {
    name        = "cluster_name"
    description = "The name of the EKS cluster. Will use the install ID by default."
    default     = ""
    required    = true
  }
}

resource "nuon_app_sandbox" "byovpc" {
  app_id            = nuon_app.byovpc.id
  terraform_version = "v1.6.3"

  public_repo = {
    repo      = "nuonco/sandboxes"
    branch    = "main"
    directory = "aws-eks-byovpc"
  }

  var {
    name  = "eks_version"
    value = "{{.nuon.install.inputs.eks_version}}"
  }

  var {
    name  = "vpc_id"
    value = "{{.nuon.install.inputs.vpc_id}}"
  }

  var {
    name  = "cluster_name"
    value = "{{.nuon.install.inputs.cluster_name}}"
  }
}

resource "nuon_install" "byovpc" {
  name         = local.cluster_name
  app_id       = nuon_app.byovpc.id
  region       = local.region
  iam_role_arn = module.install_access.iam_role_arn

  input {
    name  = "eks_version"
    value = "1.28"
  }

  input {
    name  = "vpc_id"
    value = module.vpc.vpc_id
  }

  input {
    name  = "cluster_name"
    value = local.cluster_name
  }

  depends_on = [
    nuon_app_sandbox.byovpc,
  ]
}

resource "nuon_container_image_component" "byovpc" {
  name   = "public_image"
  app_id = nuon_app.byovpc.id

  public = {
    image_url = "kennethreitz/httpbin"
    tag       = "latest"
  }
}
