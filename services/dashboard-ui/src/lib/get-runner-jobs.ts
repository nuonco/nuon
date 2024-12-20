import type { TRunnerJob } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetRunnerJobs {
  runnerId: string
  orgId: string
}

export async function getRunnerJobs({
  runnerId,
  orgId,
}: IGetRunnerJobs): Promise<Array<TRunnerJob>> {
  const data = await fetch(
    `${API_URL}/v1/runners/${runnerId}/jobs`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch runner jobs')
  }

  return data.json()
}
