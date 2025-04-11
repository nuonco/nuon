import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (
    req: NextRequest,
    { params }: TRouteRes<'org-id' | 'app-id' | 'input-id'>
  ) => {
    const orgId = params?.['org-id']
    const appId = params?.['app-id']
    const inputsId = params?.['input-id']

    return fetch(
      `${API_URL}/v1/apps/${appId}/input-configs/${inputsId}`,
      await getFetchOpts(orgId, 10000)
    ).then((res) => {
      return res.json().then((input) => {
        return NextResponse.json(input, {
          status: 200,
        })
      })
    })
  }
)
