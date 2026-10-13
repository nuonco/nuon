import { api } from '@/lib/api'
import type { TInstallOverview } from '@/types'

export const getInstallOverview = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TInstallOverview>({
    orgId,
    path: `installs/${installId}/overview`,
  })
