import { api } from '@/lib/api'
import type { TInstallActionRun } from '@/types'

export const getInstallActionRunById = ({
  installId,
  runId,
  orgId,
}: {
  installId: string
  runId: string
  orgId: string
}) =>
  api<TInstallActionRun>({
    path: `installs/${installId}/action-workflows/runs/${runId}`,
    orgId,
  })
