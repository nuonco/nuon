import { api } from '@/lib/api'
import type { TInstallAppConfigVersion } from '@/types'

export const getInstallAppConfigVersions = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TInstallAppConfigVersion[]>({
    path: `installs/${installId}/app-config-versions`,
    orgId,
  })
