import { api } from '@/lib/api'
import type { TApp } from '@/types'

export const getApp = ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) =>
  api<TApp>({
    path: `apps/${appId}`,
    orgId,
  })
