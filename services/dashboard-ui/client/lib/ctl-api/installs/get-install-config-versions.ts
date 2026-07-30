import { api } from '@/lib/api'
import type { TInstallConfigVersion } from '@/types/ctl-api.types'

export const getInstallConfigVersions = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TInstallConfigVersion[]>({
    path: `installs/${installId}/config-versions`,
    orgId,
  })
