import { api } from '@/lib/api'
import type { TInstallStatus } from '@/types'

export const getInstallStatus = ({
  orgId,
  installId,
}: {
  orgId: string
  installId: string
}) =>
  api<TInstallStatus>({
    orgId,
    path: `installs/${installId}/status`,
  })
