import { api } from '@/lib/api'
import type { TRunnerHeartbeat } from '@/types'

export const getProcessLatestHeartbeat = ({
  runnerId,
  processId,
  orgId,
}: {
  runnerId: string
  processId: string
  orgId: string
}) =>
  api<TRunnerHeartbeat>({
    path: `runners/${runnerId}/processes/${processId}/heart-beats/latest`,
    orgId,
  })
