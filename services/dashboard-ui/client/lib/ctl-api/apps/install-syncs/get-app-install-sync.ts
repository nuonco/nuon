import { api } from '@/lib/api'
import type { TAppInstallConfigSync } from '@/types'

export const getAppInstallSync = ({
  appId,
  syncId,
  orgId,
}: {
  appId: string
  syncId: string
  orgId: string
}) =>
  api<TAppInstallConfigSync>({
    path: `apps/${appId}/install-syncs/${syncId}`,
    orgId,
  })
