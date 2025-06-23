import { NextRequest, NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'workflow-id' | 'step-id'>
) => {
  const {
    ['org-id']: orgId,
    ['workflow-id']: workflowId,
    ['step-id']: stepId,
  } = await params

  const res = await nueQueryData({
    orgId,
    path: `install-workflows/${workflowId}/steps/${stepId}`,
  })

  return NextResponse.json(res)
}
