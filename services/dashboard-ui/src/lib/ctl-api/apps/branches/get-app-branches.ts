import { api } from '@/lib/api'
import type { TAppBranch } from '@/types'

export const getAppBranches = ({
  appId,
  orgId,
  limit,
  offset,
}: {
  appId: string
  orgId: string
  limit?: number
  offset?: number
}) =>
  api<TAppBranch[]>({
    path: `apps/${appId}/branches`,
    orgId,
  })
