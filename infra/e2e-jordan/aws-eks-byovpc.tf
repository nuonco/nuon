locals {
  aws_eks_byovpc_app_name    = "${local.name}-aws-eks-byovpc"
  aws_eks_byovpc_sandbox_dir = "aws-eks-byovpc"
  aws_eks_byovpc_install_inputs = [
    {
      name          = "vpc_id"
      description   = "The VPC to deploy the app to."
      default       = ""
      required      = true
      value         = module.vpc.vpc_id
      interpolation = "{{.nuon.install.inputs.vpc_id}}"
    },
    {
      name          = "cluster_name"
      description   = "The name of the EKS cluster. Will use the install ID by default."
      default       = ""
      required      = true
      value         = local.aws_eks_byovpc_app_name
      interpolation = "{{.nuon.install.inputs.cluster_name}}"
    },
    {
      name          = "eks_version"
      description   = "The Kubernetes version to use for the EKS cluster."
      default       = ""
      required      = true
      value         = "1.28"
      interpolation = "{{.nuon.install.inputs.eks_version}}"
    },
  ]
}

module "aws-eks-byovpc" {
  source = "./e2e"

  app_name = local.aws_eks_byovpc_app_name

  sandbox_repo   = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir    = local.aws_eks_byovpc_sandbox_dir

  east_1_count = 0
  east_2_count = 0
  west_2_count = 0

  install_role_arn = module.install_access.iam_role_arn

  install_inputs = local.aws_eks_byovpc_install_inputs
}

module "aws-eks-byovpc-sandbox" {
  source = "./e2e"
  providers = {
    nuon = nuon.sandbox
  }

  app_name = "${local.aws_eks_byovpc_app_name}-sandbox"

  sandbox_repo   = local.sandboxes_repo
  sandbox_branch = local.sandboxes_branch
  sandbox_dir    = local.aws_eks_byovpc_sandbox_dir

  east_1_count = 5
  east_2_count = 5
  west_2_count = 5

  install_role_arn = module.install_access.iam_role_arn

  install_inputs = local.aws_eks_byovpc_install_inputs
}
