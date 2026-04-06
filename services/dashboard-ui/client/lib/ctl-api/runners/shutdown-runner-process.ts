import { api } from '@/lib/api'
import type { TRunnerProcessShutdown } from '@/types'

export const shutdownRunnerProcess = ({
  runnerId,
  processId,
  shutdownType,
  orgId,
}: {
  runnerId: string
  processId: string
  shutdownType: 'graceful' | 'force' | 'restart'
  orgId: string
}) =>
  api<TRunnerProcessShutdown>({
    path: `runners/${runnerId}/processes/${processId}/shutdown`,
    method: 'POST',
    orgId,
    body: { shutdown_type: shutdownType },
  })
