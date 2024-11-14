import type { TLogStream } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetLogStream {
  logStreamId: string
  orgId: string
}

export async function getLogStream({
  logStreamId,
  orgId,
}: IGetLogStream): Promise<TLogStream> {
  const data = await fetch(
    `${API_URL}/v1/log-streams/${logStreamId}`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch log stream')
  }

  return data.json()
}
