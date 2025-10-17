locals {
  karpenter = {
    cluster_name    = local.workspace_trimmed
    namespace       = "kube-system"
    version         = "1.0.6"
    discovery_key   = "karpenter.sh/discovery"
    discovery_value = local.workspace_trimmed
  }
}

module "karpenter" {
  source = "./karpenter"

  cluster_name         = local.karpenter.cluster_name
  cluster_endpoint     = module.eks.cluster_endpoint
  namespace            = local.karpenter.namespace
  karpenter_version    = local.karpenter.version
  discovery_key        = local.karpenter.discovery_key
  discovery_value      = local.karpenter.discovery_value
  node_iam_role_arn    = module.eks.eks_managed_node_groups["karpenter"].iam_role_arn
  node_iam_role_name   = module.eks.eks_managed_node_groups["karpenter"].iam_role_name
  oidc_provider_arn    = module.eks.oidc_provider_arn
  tags                 = local.tags
  ec2nodeclasses       = lookup(local.vars, "ec2nodeclasses", [])
  instance_types       = local.vars.karpenter.instance_types
  additional_nodepools = local.vars.additional_nodepools
}
