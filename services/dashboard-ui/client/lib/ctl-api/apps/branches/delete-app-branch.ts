import { api } from '@/lib/api'

export const deleteAppBranch = ({
  appId,
  branchId,
  orgId,
}: {
  appId: string
  branchId: string
  orgId: string
}) =>
  api({
    method: 'DELETE',
    path: `apps/${appId}/branches/${branchId}`,
    orgId,
  })
