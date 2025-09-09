import { NextRequest, NextResponse } from 'next/server'
import { API_URL } from '@/configs/api'
import { getFetchOpts } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'log-stream-id'>
) => {
  const { ['org-id']: orgId, ['log-stream-id']: logStreamId } = await params
  const offset = req?.headers?.get('x-nuon-api-offset') || '0'
  const fetchOpts = await getFetchOpts(
    orgId,
    { 'X-Nuon-API-Offset': offset },
    10000
  )

  return fetch(`${API_URL}/v1/log-streams/${logStreamId}/logs`, fetchOpts).then(
    (res) => {
      const next = res?.headers?.get('x-nuon-api-next') || '0'
      return res.json().then((logs) => {
        return NextResponse.json(logs, {
          status: 200,
          headers: {
            'X-Nuon-API-Next': next,
          },
        })
      })
    }
  )
}
