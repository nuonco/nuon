import type { TRunnerJob } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetRunnerJob {
  jobId: string
  orgId: string
}

export async function getRunnerJob({
  jobId,
  orgId,
}: IGetRunnerJob): Promise<TRunnerJob> {
  const data = await fetch(
    `${API_URL}/v1/runner-jobs/${jobId}`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch runner job')
  }

  return data.json()
}
