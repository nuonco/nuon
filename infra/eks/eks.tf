locals {
  cluster_version = "1.30"
  region          = local.vars.region

  # rearrange SSO roles by name for easier access
  sso_roles = { for i, name in data.aws_iam_roles.sso_roles.names : tolist(split("_", name))[1] => name }
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

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

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

  manage_aws_auth_configmap = true

  aws_auth_roles = concat([
    {
      rolearn  = aws_iam_role.github_actions.arn
      username = "gha:{{SessionName}}"
      groups = [
        "system:masters",
      ]
    },
    {
      rolearn  = "arn:aws:iam::${local.target_account_id}:role/${local.sso_roles["NuonAdmin"]}"
      username = "admin:{{SessionName}}"
      groups = [
        "system:masters",
        "eks-console-dashboard-full-access",
      ]
    },
    {
      rolearn  = "arn:aws:iam::${local.target_account_id}:role/${local.sso_roles["NuonPowerUser"]}"
      username = "power-user:{{SessionName}}"
      groups = [
        "engineers",
        "eks-console-dashboard-full-access",
      ]
    },
    ],
    [
      for add in local.vars.auth_map_additions : {
        rolearn : module.extra_auth_map[add.name].iam_role_arn
        username : add.name
        groups : add.groups
      }
    ]
  )

  eks_managed_node_groups = {
    karpenter = {
      instance_types = local.vars.managed_node_group.instance_types

      min_size     = local.vars.managed_node_group.min_size
      max_size     = local.vars.managed_node_group.max_size
      desired_size = local.vars.managed_node_group.desired_size

      iam_role_additional_policies = {
        # Required by Karpenter
        ssm = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
      }
      taints = {
        no_schedule_karpenter = {
          key    = "CriticalAddonsOnly"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      }
      tags = {
        "karpenter.sh/discovery" = local.karpenter.discovery_value
      }
    }
  }

  # HACK: https://github.com/terraform-aws-modules/terraform-aws-eks/issues/1986
  # NOTE(fd): i think we still need this
  # Because of this, we do a few things:
  # 1.) we don't add tags to the cluster primary security group, so it won't have the "owned" tag
  # 2.) we only add the karpenter tag on this node group
  # 3.) we add the "owned" tag on this node group
  node_security_group_tags = merge(local.tags, {
    "kubernetes.io/cluster/${local.workspace_trimmed}" = null

    # NOTE - if creating multiple security groups with this module, only tag the
    # security group that Karpenter should utilize with the following tag
    # (i.e. - at most, only one security group should have this tag in your account)
    (local.karpenter.discovery_key) = local.karpenter.discovery_value
  })
  create_cluster_primary_security_group_tags = false

  # this can't rely on default_tags.
  # full set of tags must be specified here :sob:
  tags = merge(local.tags, {
    "karpenter.sh/discovery" = local.karpenter.discovery_value
  })
}

resource "kubectl_manifest" "cluster_role_dashboard" {
  yaml_body = yamlencode({
    apiVersion = "rbac.authorization.k8s.io/v1"
    kind       = "ClusterRole"
    metadata = {
      name = "eks-console-dashboard-full-access"
    }
    rules = [
      {
        apiGroups = ["", ]
        resources = ["nodes", "namespaces", "pods",
        ]
        verbs = ["get", "list", ]
      },
      {
        apiGroups = ["apps", ]
        resources = [
          "deployments",
          "daemonsets",
          "statefulsets",
          "replicasets",
        ]
        verbs = ["get", "list", ]
      },
      {
        apiGroups = ["batch", ]
        resources = ["jobs", ]
        verbs     = ["get", "list", ]
      },
    ]
  })

  depends_on = [
    helm_release.karpenter
  ]
}

resource "kubectl_manifest" "cluster_role_binding_dashboard" {
  yaml_body = yamlencode({
    apiVersion = "rbac.authorization.k8s.io/v1"
    kind       = "ClusterRoleBinding"
    metadata = {
      name = "eks-console-dashboard-full-access"
    }
    roleRef = {
      apiGroup = "rbac.authorization.k8s.io"
      kind     = "ClusterRole"
      name     = "eks-console-dashboard-full-access"
    }
    subjects = [
      {
        apiGroup = "rbac.authorization.k8s.io"
        kind     = "Group"
        name     = "eks-console-dashboard-full-access"
      },
    ]
  })

  depends_on = [
    kubectl_manifest.cluster_role_dashboard
  ]
}

resource "kubectl_manifest" "cluster_role_binding_engineers_edit" {
  yaml_body = yamlencode({
    apiVersion = "rbac.authorization.k8s.io/v1"
    kind       = "ClusterRoleBinding"
    metadata = {
      name = "engineers-edit"
    }
    roleRef = {
      apiGroup = "rbac.authorization.k8s.io"
      kind     = "ClusterRole"
      name     = "edit"
    }
    subjects = [
      {
        apiGroup = "rbac.authorization.k8s.io"
        kind     = "Group"
        name     = "engineers"
      },
    ]
  })

  depends_on = [
    helm_release.karpenter
  ]
}
