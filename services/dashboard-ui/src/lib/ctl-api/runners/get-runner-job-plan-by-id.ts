import { api } from '@/lib/api'
import type { TRunnerJobPlan } from '@/types'

export const getRunnerJobPlanById = ({
  runnerJobId,
  orgId,
}: {
  runnerJobId: string
  orgId: string
}) =>
  api<TRunnerJobPlan>({
    path: `runner-jobs/${runnerJobId}/plan`,
    orgId,
  })
