locals {
  cluster_version = local.vars.cluster_version
  region          = local.vars.region

  # rearrange SSO roles by name for easier access
  sso_roles = { for i, name in data.aws_iam_roles.sso_roles.names : tolist(split("_", name))[1] => name }

  # custom entries used for external access from other clusters
  external_access_entries = { for idx, item in local.vars.auth_map_additions : format("external-access-%s", item.name) => {
    principal_arn     = module.extra_auth_map[item.name].iam_role_arn
    kubernetes_groups = [] # empty because they are all system:admin and those are replaced by AmazonEKSClusterAdminPolicy
    policy_associations = {
      cluster_admin = {
        policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
        access_scope = {
          type = "cluster"
        }
      }
    }
  } }

  # default access entries used for each cluster
  default_access_entries = {
    "gha:{{SessionName}}" = {
      principal_arn = aws_iam_role.github_actions.arn
      kubernetes_groups = [
        # "system:masters",
      ],
      policy_associations = {
        cluster_admin = {
          policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
          access_scope = {
            type = "cluster"
          }
        }
      }
    }
    "admin:{{SessionName}}" = {
      # principal_arn = "arn:aws:iam::${local.target_account_id}:role/${local.sso_roles["NuonAdmin"]}"
      # trying based on this: https://github.com/terraform-aws-modules/terraform-aws-eks/issues/2969
      #                                                                                                         hardcoded ⤵
      principal_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-reserved/sso.amazonaws.com/us-east-2/${local.sso_roles["NuonAdmin"]}"
      kubernetes_groups = [
        # "system:masters",
        "eks-console-dashboard-full-access",
      ]
      policy_associations = {
        cluster_admin = {
          policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
          access_scope = {
            type = "cluster"
          }
        }
      }
    },
    "power-user:{{SessionName}}" = {
      # principal_arn = "arn:aws:iam::${local.target_account_id}:role/${local.sso_roles["NuonPowerUser"]}"
      # trying based on this: https://github.com/terraform-aws-modules/terraform-aws-eks/issues/2969
      #                                                                                                         hardcoded ⤵
      principal_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-reserved/sso.amazonaws.com/us-east-2/${local.sso_roles["NuonPowerUser"]}"
      kubernetes_groups = [
        "engineers",
        "eks-console-dashboard-full-access",
      ]
      policy_associations = {
        cluster_admin = {
          policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSAdminViewPolicy"
          access_scope = {
            type = "cluster"
          }
        }
      }
    }
  }
}

# the role names are like AWSReservedSSO_${ROLE}_${random_stuff}
data "aws_iam_roles" "sso_roles" {
  name_regex  = "AWSReservedSSO_*"
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}

resource "aws_kms_key" "eks" {
  description = "Key for ${local.workspace_trimmed} EKS cluster"
}

resource "aws_kms_alias" "eks" {
  name          = "alias/nuon/eks-${local.workspace_trimmed}"
  target_key_id = aws_kms_key.eks.id
}

# https://github.com/terraform-aws-modules/terraform-aws-eks/blob/master/docs/UPGRADE-20.0.md
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "20.37.0"

  # This module does something funny with state and `default_tags`
  # so it shows as a change on every apply. By using a provider w/o
  # `default_tags`, we can avoid this?
  providers = {
    aws = aws.no_tags
  }

  cluster_name                    = local.workspace_trimmed
  cluster_version                 = local.cluster_version
  cluster_endpoint_private_access = true
  cluster_endpoint_public_access  = true

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets


  authentication_mode = "API_AND_CONFIG_MAP"
  enable_irsa         = true

  # Cluster access entry
  # https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest#cluster-access-entry
  # https://docs.aws.amazon.com/eks/latest/userguide/access-policies.html
  enable_cluster_creator_admin_permissions = true # aws-auth configmap-based auth for terraform
  access_entries = merge(
    local.default_access_entries,
    local.external_access_entries
  )

  create_kms_key = false
  cluster_encryption_config = {
    provider_key_arn = aws_kms_key.eks.arn
    resources        = ["secrets"]
  }

  node_security_group_additional_rules = {
    ingress_self_all = {
      description = "Node to node all ports/protocols"
      protocol    = "-1"
      from_port   = 0
      to_port     = 0
      type        = "ingress"
      self        = true
    }

    egress_all = {
      description      = "Node all egress"
      protocol         = "-1"
      from_port        = 0
      to_port          = 0
      type             = "egress"
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = ["::/0"]
    }
  }

  cluster_addons = {
    coredns = {
      configuration_values = jsonencode({
        tolerations = [
          # Allow CoreDNS to run on the same nodes as the Karpenter controller
          # for use during cluster creation when Karpenter nodes do not yet exist
          {
            key    = "karpenter.sh/controller"
            value  = "true"
            effect = "NoSchedule"
          }
        ]
      })
    }
    eks-pod-identity-agent = {}
  }

  eks_managed_node_groups = {
    karpenter = {
      instance_types = local.vars.managed_node_group.instance_types

      min_size     = local.vars.managed_node_group.min_size
      max_size     = local.vars.managed_node_group.max_size
      desired_size = local.vars.managed_node_group.desired_size

      # TODO(fd): idk if we still need this
      # iam_role_additional_policies = {
      #   # Required by Karpenter
      #   ssm = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
      # }

      # Used to ensure Karpenter runs on nodes that it does not manage
      labels = {
        "karpenter.sh/controller" = "true"
      }
      # won't schedule on nodes it manages
      taints = {
        karpenter = {
          key    = "karpenter.sh/controller"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      }
      tags = {
        "karpenter.sh/discovery" = local.karpenter.discovery_value
      }
    }
  }

  create_cluster_primary_security_group_tags = false
  node_security_group_tags = merge(local.tags, {
    # NOTE - if creating multiple security groups with this module, only tag the
    # security group that Karpenter should utilize with the following tag
    # (i.e. - at most, only one security group should have this tag in your account)
    "karpenter.sh/discovery" = local.karpenter.discovery_value
  })

  # this can't rely on default_tags.
  # full set of tags must be specified here :sob:
  tags = merge(local.tags, {
    "karpenter.sh/discovery" = local.karpenter.discovery_value
  })
}

# service linked roles use the aws-auth configmap-based auth
# until they are valid principals for an entry
module "eks_aws_auth" {
  source  = "terraform-aws-modules/eks/aws//modules/aws-auth"
  version = "~> 20.0"

  manage_aws_auth_configmap = true

  # this needs to be converted to entries
  aws_auth_roles = [
    {
      username = "admin:{{SessionName}}"
      # trying based on this: https://github.com/terraform-aws-modules/terraform-aws-eks/issues/2969
      rolearn = "arn:aws:iam::${local.target_account_id}:role/${local.sso_roles["NuonAdmin"]}"
      groups = [
        "system:masters",
        "eks-console-dashboard-full-access",
      ]
    },
    {
      username = "power-user:{{SessionName}}"
      rolearn  = "arn:aws:iam::${local.target_account_id}:role/${local.sso_roles["NuonPowerUser"]}"
      groups = [
        "engineers",
        "eks-console-dashboard-full-access",
      ]
    }
  ]
}

# remove from state manually
# resource "kubectl_manifest" "cluster_role_dashboard" {
#   yaml_body = yamlencode({
#     apiVersion = "rbac.authorization.k8s.io/v1"
#     kind       = "ClusterRole"
#     metadata = {
#       name = "eks-console-dashboard-full-access"
#     }
#     rules = [
#       {
#         apiGroups = ["", ]
#         resources = ["nodes", "namespaces", "pods",
#         ]
#         verbs = ["get", "list", ]
#       },
#       {
#         apiGroups = ["apps", ]
#         resources = [
#           "deployments",
#           "daemonsets",
#           "statefulsets",
#           "replicasets",
#         ]
#         verbs = ["get", "list", ]
#       },
#       {
#         apiGroups = ["batch", ]
#         resources = ["jobs", ]
#         verbs     = ["get", "list", ]
#       },
#     ]
#   })

#   depends_on = [
#     module.eks_aws_auth
#   ]
# }

# resource "kubectl_manifest" "cluster_role_binding_dashboard" {
#   yaml_body = yamlencode({
#     apiVersion = "rbac.authorization.k8s.io/v1"
#     kind       = "ClusterRoleBinding"
#     metadata = {
#       name = "eks-console-dashboard-full-access"
#     }
#     roleRef = {
#       apiGroup = "rbac.authorization.k8s.io"
#       kind     = "ClusterRole"
#       name     = "eks-console-dashboard-full-access"
#     }
#     subjects = [
#       {
#         apiGroup = "rbac.authorization.k8s.io"
#         kind     = "Group"
#         name     = "eks-console-dashboard-full-access"
#       },
#     ]
#   })

#   depends_on = [
#     kubectl_manifest.cluster_role_dashboard
#   ]
# }

# resource "kubectl_manifest" "cluster_role_binding_engineers_edit" {
#   yaml_body = yamlencode({
#     apiVersion = "rbac.authorization.k8s.io/v1"
#     kind       = "ClusterRoleBinding"
#     metadata = {
#       name = "engineers-edit"
#     }
#     roleRef = {
#       apiGroup = "rbac.authorization.k8s.io"
#       kind     = "ClusterRole"
#       name     = "edit"
#     }
#     subjects = [
#       {
#         apiGroup = "rbac.authorization.k8s.io"
#         kind     = "Group"
#         name     = "engineers"
#       },
#     ]
#   })

#   depends_on = [
#     module.eks_aws_auth
#   ]
# }
