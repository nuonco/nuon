import { NextRequest, NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'app-id'>
) => {
  const { ['org-id']: orgId, ['app-id']: appId } = await params

  const res = await nueQueryData({
    orgId,
    path: `apps/${appId}/latest-config`,
  })

  return NextResponse.json(res)
}
