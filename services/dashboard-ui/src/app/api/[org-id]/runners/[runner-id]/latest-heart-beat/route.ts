import { NextRequest, NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'runner-id'>
) => {
  const { ['org-id']: orgId, ['runner-id']: runnerId } = await params

  const res = await nueQueryData({
    orgId,
    path: `runners/${runnerId}/latest-heart-beat`,
  })

  return NextResponse.json(res)
}
