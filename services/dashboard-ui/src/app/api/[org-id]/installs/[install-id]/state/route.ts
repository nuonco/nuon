import { NextRequest, NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'install-id'>
) => {
  const { ['org-id']: orgId, ['install-id']: installId } = await params

  const res = await nueQueryData({
    orgId,
    path: `installs/${installId}/state`,
  })

  return NextResponse.json(res)
}
