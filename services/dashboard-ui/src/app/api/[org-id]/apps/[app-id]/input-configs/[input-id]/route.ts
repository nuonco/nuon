import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, appId, __, inputsId] = req.url.split('/').slice(4, 9)

  return fetch(
    `${API_URL}/v1/apps/${appId}/input-configs`,
    await getFetchOpts(orgId, 10000)
  ).then((res) => {
    return res.json().then((inputs) => {
      const input = inputs?.find((inp) => inp?.id === inputsId)
      return NextResponse.json(input, {
        status: 200,
      })
    })
  })
})
