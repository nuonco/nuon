import { api } from '@/lib/api'
import type { TRunnerJob } from '@/types'

export const getRunnerJob = ({
  runnerJobId,
  orgId,
}: {
  runnerJobId: string
  orgId: string
}) =>
  api<TRunnerJob>({
    path: `runner-jobs/${runnerJobId}`,
    orgId,
  })
