import { api } from '@/lib/api'
import type { TAppBranchConfig } from '@/types'

// The branch's newest VCS/install-group config (404 when it has none), not an
// app config.
export const getBranchLatestConfig = ({
  appId,
  branchId,
  orgId,
}: {
  appId: string
  branchId: string
  orgId: string
}) =>
  api<TAppBranchConfig>({
    path: `apps/${appId}/branches/${branchId}/latest-config`,
    orgId,
  })
