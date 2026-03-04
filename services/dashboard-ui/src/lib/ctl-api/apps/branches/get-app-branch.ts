import { api } from '@/lib/api'
import type { TAppBranch } from '@/types'

export const getAppBranch = ({
  appId,
  branchId,
  orgId,
  latestConfig = false,
}: {
  appId: string
  branchId: string
  orgId: string
  latestConfig?: boolean
}) =>
  api<TAppBranch>({
    path: `apps/${appId}/branches/${branchId}${latestConfig ? '?latest_config=true' : ''}`,
    orgId,
  })