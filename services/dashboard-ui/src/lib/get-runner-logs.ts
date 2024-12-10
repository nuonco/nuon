import type { TOTELLog } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

// TODO(nnnnat): remove all runner logs code

export interface IGetRunnerLogs {
  runnerId: string
  jobId: string
  orgId: string
}

export async function getRunnerLogs({
  jobId,
  runnerId,
  orgId,
}: IGetRunnerLogs): Promise<Array<TOTELLog>> {
  const data = await fetch(
    `${API_URL}/v1/runners/${runnerId}/logs?job_id=${jobId}`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch runner logs')
  }

  return data.json()
}
