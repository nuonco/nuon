import { api } from '@/lib/api'
import type { TInstallReconciliation } from '@/types'

export const getInstallReconciliation = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TInstallReconciliation>({
    orgId,
    path: `installs/${installId}/reconciliation`,
  })
