output "all" {
  description = "karpenter outputs"
  value = {
    namespace               = module.karpenter.namespace
    service_account         = module.karpenter.service_account
    iam_role_arn            = module.karpenter.iam_role_arn
    iam_role_name           = module.karpenter.iam_role_name
    iam_role_unique_id      = module.karpenter.iam_role_unique_id
    instance_profile_arn    = module.karpenter.instance_profile_arn
    instance_profile_id     = module.karpenter.instance_profile_id
    instance_profile_name   = module.karpenter.instance_profile_name
    instance_profile_unique = module.karpenter.instance_profile_unique
    node_access_entry_arn   = module.karpenter.node_access_entry_arn
    node_iam_role_arn       = module.karpenter.node_iam_role_arn
    node_iam_role_name      = module.karpenter.node_iam_role_name
    node_iam_role_unique_id = module.karpenter.node_iam_role_unique_id
    queue_arn               = module.karpenter.queue_arn
    queue_name              = module.karpenter.queue_name
    queue_url               = module.karpenter.queue_url
    event_rules             = module.karpenter.event_rules
  }
}
