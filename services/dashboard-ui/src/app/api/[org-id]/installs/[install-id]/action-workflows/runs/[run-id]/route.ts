import { NextRequest, NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'install-id' | 'run-id'>
) => {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['run-id']: runId,
  } = await params

  const res = await nueQueryData({
    orgId,
    path: `installs/${installId}/action-workflows/runs/${runId}`,
  })

  return NextResponse.json(res)
}
