import { api } from '@/lib/api'
import type { TComponentBuild } from '@/types'

export const getBranchRunBuilds = ({
  appId,
  branchId,
  runId,
  orgId,
}: {
  appId: string
  branchId: string
  runId: string
  orgId: string
}) =>
  api<TComponentBuild[]>({
    path: `apps/${appId}/branches/${branchId}/runs/${runId}/builds`,
    orgId,
  })
