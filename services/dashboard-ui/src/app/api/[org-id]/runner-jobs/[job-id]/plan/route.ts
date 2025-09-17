import { NextRequest, NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'job-id'>
) => {
  const { ['org-id']: orgId, ['job-id']: runnerJobId } = await params

  const res = await nueQueryData({
    orgId,
    path: `runner-jobs/${runnerJobId}/plan`,
  })
  
  return NextResponse.json(res)
}
