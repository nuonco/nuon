import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, appId, __, configId] = req.url.split('/').slice(4, 9)

  return fetch(
    `${API_URL}/v1/apps/${appId}/sandbox-configs`,
    await getFetchOpts(orgId, 10000)
  ).then((res) => {
    return res.json().then((sandboxs) => {
      const sandbox =
        configId === 'latest'
          ? sandboxs[0]
          : sandboxs?.find((sbx) => sbx?.id === configId)
      return NextResponse.json(sandbox, {
        status: 200,
      })
    })
  })
})
