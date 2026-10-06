import { api } from '@/lib/api'
import type { TAppConfigDiffResponse } from '@/lib/ctl-api/apps/get-app-config-diff'

export const getInstallAppConfigTreeDiff = ({
  installId,
  configId,
  orgId,
}: {
  installId: string
  configId: string
  orgId: string
}) =>
  api<TAppConfigDiffResponse>({
    path: `installs/${installId}/app-configs/${configId}/diff`,
    orgId,
  })
