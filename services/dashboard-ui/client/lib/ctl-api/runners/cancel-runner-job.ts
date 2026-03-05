import { api } from '@/lib/api'
import type { TRunnerJob } from '@/types'

export async function cancelRunnerJob({
  orgId,
  runnerJobId,
}: {
  runnerJobId: string
  orgId: string
}) {
  return api<TRunnerJob>({
    method: 'POST',
    orgId,
    path: `runner-jobs/${runnerJobId}/cancel`,
  })
}
