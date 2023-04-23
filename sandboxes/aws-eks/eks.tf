locals {
  cluster_version = "1.24"
  region          = local.vars.region
}

resource "aws_kms_key" "eks" {
  description = "Key for ${local.vars.id} EKS cluster"
}

resource "aws_kms_alias" "eks" {
  name          = "alias/nuon/eks-${local.vars.id}"
  target_key_id = aws_kms_key.eks.id
}

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  # This module does something funny with state and `default_tags`
  # so it shows as a change on every apply. By using a provider w/o
  # `default_tags`, we can avoid this?
  providers = {
    aws = aws.no_tags
  }

  cluster_name                    = local.vars.id
  cluster_version                 = local.cluster_version
  cluster_endpoint_private_access = true
  # TODO(jdt): This seems prudent to set to false
  cluster_endpoint_public_access = true

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  create_kms_key = false
  cluster_encryption_config = {
    provider_key_arn = aws_kms_key.eks.arn
    resources        = ["secrets"]
  }

  node_security_group_additional_rules = {}

  manage_aws_auth_configmap = true

  # TODO(jdt): These shouldn't be `system:masters`
  # We should create a role with the correct permissions instead of giving "god mode"
  aws_auth_roles = [
    # Allow the install role to have access to create resources
    {
      rolearn  = var.install_role_arn
      username = "install:{{SessionName}}"
      groups = [
        "system:masters",
      ]
    },
    # Allow for updates via terraform
    {
      rolearn  = var.assume_role_arn
      username = "terraform:{{SessionName}}"
      groups = [
        "system:masters",
      ]
    },

  ]

  eks_managed_node_groups = {
    default = {
      instance_types = local.vars.instance_types

      min_size     = local.vars.min_size
      max_size     = local.vars.max_size
      desired_size = local.vars.desired_size

      iam_role_additional_policies = {
        additional = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
      }
    }
  }

  # HACK: https://github.com/terraform-aws-modules/terraform-aws-eks/issues/1986
  node_security_group_tags = {
    "kubernetes.io/cluster/${local.vars.id}" = null
  }

  # this can't rely on default_tags.
  # full set of tags must be specified here :sob:
  tags = local.tags
}
