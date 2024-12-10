import type { TOTELLog } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

async function pageLogs({ orgId, logStreamId, next = '0' }) {
  return fetch(
    `${API_URL}/v1/log-streams/${logStreamId}/logs`,
    await getFetchOpts(orgId, { 'X-Nuon-API-Offset': next })
  ).then(async (res) => {
    if (res.ok) {
      if (res.headers.get('x-nuon-api-next')) {
        return [
          ...(await res.json()),
          ...(await pageLogs({
            orgId,
            logStreamId,
            next: res.headers.get('x-nuon-api-next'),
          })),
        ]
      } else {
        return res.json()
      }
    } else {
      throw new Error('Failed to fetch log stream logs')
    }
  })
}

export interface IGetLogStreamLogs {
  logStreamId: string
  orgId: string
}

export async function getLogStreamLogs({
  logStreamId,
  orgId,
}: IGetLogStreamLogs): Promise<Array<TOTELLog>> {
  return pageLogs({ logStreamId, orgId })
}
