import { NextRequest, NextResponse } from 'next/server'
import { API_URL } from '@/configs/api'
import { getFetchOpts } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'app-id' | 'input-id'>
) => {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['input-id']: inputsId,
  } = await params

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
