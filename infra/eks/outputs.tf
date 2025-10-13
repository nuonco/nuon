output "cluster_arn" {
  description = "The Amazon Resource Name (ARN) of the cluster"
  value       = module.eks.cluster_arn
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data required to communicate with the cluster"
  value       = module.eks.cluster_certificate_authority_data
}

output "cluster_endpoint" {
  description = "Endpoint for your Kubernetes API server"
  value       = module.eks.cluster_endpoint
}

output "cluster_id" {
  description = <<EOT
  DEPRECATED: The ID used to be the name. Use `cluster_name`
  With Outposts, the ID is a UUID/GUID and the underlying module re-used ID and created name.
  The name of the EKS cluster. Will block on cluster creation until the cluster is really ready"
  EOT
  value       = module.eks.cluster_name
}

output "cluster_name" {
  description = "The name of the EKS cluster. Will block on cluster creation until the cluster is really ready"
  value       = module.eks.cluster_name
}

output "cluster_oidc_issuer_url" {
  description = "The URL on the EKS cluster for the OpenID Connect identity provider"
  value       = module.eks.cluster_oidc_issuer_url
}

output "cluster_platform_version" {
  description = "Platform version for the cluster"
  value       = module.eks.cluster_platform_version
}

output "cluster_status" {
  description = "Status of the EKS cluster. One of `CREATING`, `ACTIVE`, `DELETING`, `FAILED`"
  value       = module.eks.cluster_status
}

output "cluster_primary_security_group_id" {
  description = "Cluster security group that was created by Amazon EKS for the cluster. Managed node groups use this security group for control-plane-to-data-plane communication. Referred to as 'Cluster security group' in the EKS console"
  value       = module.eks.cluster_primary_security_group_id
}

################################################################################
# Security Group
################################################################################

output "cluster_security_group_arn" {
  description = "Amazon Resource Name (ARN) of the cluster security group"
  value       = module.eks.cluster_security_group_arn
}

output "cluster_security_group_id" {
  description = "ID of the cluster security group"
  value       = module.eks.cluster_security_group_id
}

################################################################################
# Node Security Group
################################################################################

output "node_security_group_arn" {
  description = "Amazon Resource Name (ARN) of the node shared security group"
  value       = module.eks.node_security_group_arn
}

output "node_security_group_id" {
  description = "ID of the node shared security group"
  value       = module.eks.node_security_group_id
}

################################################################################
# IRSA
################################################################################

output "oidc_provider" {
  description = "The OpenID Connect identity provider (issuer URL without leading `https://`)"
  value       = module.eks.oidc_provider
}

output "oidc_provider_arn" {
  description = "The ARN of the OIDC Provider if `enable_irsa = true`"
  value       = module.eks.oidc_provider_arn
}

################################################################################
# IAM Role
################################################################################

output "cluster_iam_role_name" {
  description = "IAM role name of the EKS cluster"
  value       = module.eks.cluster_iam_role_name
}

output "cluster_iam_role_arn" {
  description = "IAM role ARN of the EKS cluster"
  value       = module.eks.cluster_iam_role_arn
}

output "cluster_iam_role_unique_id" {
  description = "Stable and unique string identifying the IAM role"
  value       = module.eks.cluster_iam_role_unique_id
}

################################################################################
# EKS Addons
################################################################################

output "cluster_addons" {
  description = "List of EKS cluster addons."
  value       = [for addon in module.eks.cluster_addons : addon.arn]
}


################################################################################
# EKS Identity Provider
################################################################################

output "cluster_identity_providers" {
  description = "Map of attribute maps for all EKS identity providers enabled"
  value       = module.eks.cluster_identity_providers
}

################################################################################
# CloudWatch Log Group
################################################################################

output "cloudwatch_log_group_name" {
  description = "Name of cloudwatch log group created"
  value       = module.eks.cloudwatch_log_group_name
}

output "cloudwatch_log_group_arn" {
  description = "Arn of cloudwatch log group created"
  value       = module.eks.cloudwatch_log_group_arn
}

################################################################################
# Fargate Profile
################################################################################

output "fargate_profiles" {
  description = "Map of attribute maps for all EKS Fargate Profiles created"
  value       = module.eks.fargate_profiles
}

################################################################################
# EKS Managed Node Group
################################################################################

output "eks_managed_node_groups" {
  description = "Map of attribute maps for all EKS managed node groups created"
  value       = module.eks.eks_managed_node_groups
}

output "eks_managed_node_groups_autoscaling_group_names" {
  description = "List of the autoscaling group names created by EKS managed node groups"
  value       = module.eks.eks_managed_node_groups_autoscaling_group_names
}

################################################################################
# Self Managed Node Group
################################################################################

output "self_managed_node_groups" {
  description = "Map of attribute maps for all self managed node groups created"
  value       = module.eks.self_managed_node_groups
}

output "self_managed_node_groups_autoscaling_group_names" {
  description = "List of the autoscaling group names created by self-managed node groups"
  value       = module.eks.self_managed_node_groups_autoscaling_group_names
}

################################################################################
# Additional
################################################################################

# needs to be a map; a list of string breaks the parser
output "access_entries" {
  # we avoid printing the whole thing because the kubernetes_groups list of strings breaks the parser on gha
  description = "Access Entries' ARNs"
  value       = [for entry in module.eks.access_entries : entry.access_entry_arn]
}

output "private_zone" {
  description = "The subdomain used as the private zone for this cluster"
  value       = local.dns.zone
}

output "github_action_role_arn" {
  description = "The ARN of the role to be assumed by Github Actions"
  value       = aws_iam_role.github_actions.arn
}

output "cluster_gh_role_arn" {
  description = "The ARN of the role to be assumed by Github Actions"
  value       = aws_iam_role.github_actions.arn
}

output "auth_map_additional_role_arns" {
  description = "The ARNs of the assumable roles indexed by assuming role name"
  value = {
    for add in local.vars.auth_map_additions :
    add.name => module.extra_auth_map[add.name].iam_role_arn
  }
}

output "root_domain" {
  description = "Root domain for this environment"
  value       = data.aws_route53_zone.env_root.name
}

output "twingate_service_accounts" {
  description = "twingate service account"
  value = {
    github_actions = {
      token = nonsensitive(twingate_service_account_key.github_actions.token)
      id    = twingate_service_account_key.github_actions.service_account_id
    }

    office = {
      token = nonsensitive(twingate_service_account_key.office.token)
      id    = twingate_service_account_key.office.service_account_id
    }
  }
}

################################################################################
# Karpenter
################################################################################

#output "karpenter" {
#description = "karpenter outputs"
#value = module.karpenter.all
#}
