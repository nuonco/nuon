import { api } from '@/lib/api'
import type { TLogStream } from '@/types'

export const deleteInstallActionRunLogs = ({
  installId,
  runId,
  orgId,
}: {
  installId: string
  runId: string
  orgId: string
}) =>
  api<TLogStream>({
    abortTimeout: 70000,
    method: 'DELETE',
    orgId,
    path: `installs/${installId}/actions/runs/${runId}/logs`,
  })
