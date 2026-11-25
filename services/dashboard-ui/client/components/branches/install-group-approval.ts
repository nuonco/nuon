export const installGroupApprovalLabel = (group?: {
  approval_option?: 'approve-all' | 'prompt' | string | null
  install_approval_option?: 'approve-all' | 'prompt' | string | null
  auto_approve_on_policies_passing?: boolean | null
}) => {
  if (!group) return undefined
  const plan =
    group.approval_option === 'approve-all' ||
    !!group.auto_approve_on_policies_passing
  const installs = group.install_approval_option === 'approve-all'
  if (plan && installs) return 'Auto-approves plan and installs'
  if (plan) return "Auto-approves this group's plan"
  if (installs) return 'Auto-approves installs'
  return 'Manual approval'
}
