import { api } from '@/lib/api'
import type { TAppInputConfig } from '@/types'

export const getAppInstallInputConfig = ({
  appId,
  orgId,
}: {
  orgId: string
  appId: string
}) =>
  api<TAppInputConfig>({
    path: `apps/${appId}/install-input-config`,
    orgId,
  })
