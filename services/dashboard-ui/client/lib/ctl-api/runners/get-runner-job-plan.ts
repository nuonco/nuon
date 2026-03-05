import { api } from '@/lib/api'
import type { TRunnerJobPlan } from '@/types'

export const getRunnerJobPlan = ({
  runnerJobId,
  orgId,
}: {
  runnerJobId: string
  orgId: string
}) =>
  api<TRunnerJobPlan>({
    path: `runner-jobs/${runnerJobId}/composite-plan`,
    orgId,
  })
