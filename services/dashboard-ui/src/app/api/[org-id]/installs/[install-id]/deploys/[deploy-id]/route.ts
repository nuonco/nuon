import { NextRequest, NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'install-id' | 'deploy-id'>
) => {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['deploy-id']: deployId,
  } = await params

  const res = await nueQueryData({
    orgId,
    path: `installs/${installId}/deploys/${deployId}`,
  })

  return NextResponse.json(res)
}
