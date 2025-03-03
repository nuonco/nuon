import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, logStreamId] = req.url.split('/').slice(4, 7)
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
})
