import { api } from '@/lib/api'
import type { TAppInstallsConfig } from '@/types'

export const getAppInstallsConfig = ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) =>
  api<TAppInstallsConfig>({
    path: `apps/${appId}/installs-configs`,
    orgId,
  })
