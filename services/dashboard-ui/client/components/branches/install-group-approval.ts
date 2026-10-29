export const installGroupApprovalLabel = (group?: {
  auto_approve_on_policies_passing?: boolean | null
  do_not_deploy?: boolean | null
}) => {
  if (!group) return undefined
  if (group.do_not_deploy) return undefined
  return group.auto_approve_on_policies_passing
    ? 'Auto-approves when policies pass'
    : 'Manual approval'
}
