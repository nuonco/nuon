import { api } from '@/lib/api'
import type { TAppBranch } from '@/types'

export const getAppBranch = ({
  appId,
  branchId,
  orgId,
}: {
  appId: string
  branchId: string
  orgId: string
}) =>
  api<TAppBranch>({
    path: `apps/${appId}/branches/${branchId}`,
    orgId,
  })
