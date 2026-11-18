export const installGroupApprovalLabel = (group?: {
  approval_option?: 'approve-all' | 'prompt' | string | null
}) => {
  if (!group) return undefined
  return group.approval_option === 'approve-all'
    ? "Auto-approves this group's plan"
    : 'Manual approval'
}
