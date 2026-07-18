import { api } from '@/lib/api'
import type { TInstallConfigSync } from '@/types/ctl-api.types'

export const getInstallConfigSyncs = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TInstallConfigSync[]>({
    path: `installs/${installId}/config-syncs`,
    orgId,
  })
