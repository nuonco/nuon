import { api } from '@/lib/api'
import type { TAppInstallConfigSync } from '@/types'

export const getAppInstallSyncs = ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) =>
  api<TAppInstallConfigSync[]>({
    path: `apps/${appId}/install-syncs`,
    orgId,
  })
