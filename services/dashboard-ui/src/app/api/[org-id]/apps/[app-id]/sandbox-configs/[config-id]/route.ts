import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (
    req: NextRequest,
    { params }: TRouteRes<'org-id' | 'app-id' | 'config-id'>
  ) => {
    const orgId = params?.['org-id']
    const appId = params?.['app-id']
    const configId = params?.['config-id']

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
  }
)
