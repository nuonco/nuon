import { api } from '@/lib/api'
import type { TRunnerProcess } from '@/types'

export const getRunnerProcess = ({
  runnerId,
  processId,
  orgId,
}: {
  runnerId: string
  processId: string
  orgId: string
}) =>
  api<TRunnerProcess>({
    path: `runners/${runnerId}/processes/${processId}`,
    orgId,
  })
