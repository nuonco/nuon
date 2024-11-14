import type { TOTELLog } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetLogStreamLogs {
  logStreamId: string
  orgId: string
}

export async function getLogStreamLogs({
  logStreamId,
  orgId,
}: IGetLogStreamLogs): Promise<Array<TOTELLog>> {
  const data = await fetch(
    `${API_URL}/v1/log-streams/${logStreamId}/logs`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch log stream logs')
  }

  return data.json()
}
