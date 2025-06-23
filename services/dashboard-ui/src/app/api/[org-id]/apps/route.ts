import { NextRequest, NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id'>
) => {
  const { ['org-id']: orgId } = await params

  const res = await nueQueryData({
    orgId,
    path: `apps`,
  })

  return NextResponse.json(res)
}
