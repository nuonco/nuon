import { NextRequest, NextResponse } from 'next/server'
import { API_URL } from '@/configs/api'
import { getFetchOpts } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'app-id' | 'config-id'>
) => {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['config-id']: configId,
  } = await params

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
