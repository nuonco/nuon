export const installGroupApprovalLabel = (group?: {
  auto_approve_on_policies_passing?: boolean | null
}) => {
  if (!group) return undefined
  return group.auto_approve_on_policies_passing
    ? 'Auto-approves when policies pass'
    : 'Manual approval'
}
