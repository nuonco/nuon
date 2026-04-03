import { api } from '@/lib/api'
import type { TRunnerProcess } from '@/types'

export const getCurrentRunnerProcesses = ({
  runnerId,
  orgId,
}: {
  runnerId: string
  orgId: string
}) =>
  api<TRunnerProcess[]>({
    path: `runners/${runnerId}/processes/current`,
    orgId,
  })
