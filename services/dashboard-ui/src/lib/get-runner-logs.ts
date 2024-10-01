import type { TOTELLog } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetRunnerLogs {
  runnerId: string
  orgId: string
}

export async function getRunnerLogs({ runnerId, orgId }: IGetRunnerLogs): Promise<Array<TOTELLog>> {
  const data = await fetch(
    `${API_URL}/v1/runners/${runnerId}/logs`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch runner logs')
  }

  return data.json()
}
