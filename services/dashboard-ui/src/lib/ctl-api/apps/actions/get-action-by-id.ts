import { api } from '@/lib/api'
import type { TAction } from '@/types'

export const getActionById = ({
  actionId,
  appId,
  orgId,
}: {
  actionId: string
  appId: string
  orgId: string
}) =>
  api<TAction>({
    path: `apps/${appId}/action-workflows/${actionId}`,
    orgId,
  })
