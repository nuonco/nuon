import { NextRequest, NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'workflow-id' | 'step-id' | 'approval-id'>
) => {
  const {
    ['org-id']: orgId,
    ['workflow-id']: workflowId,
    ['step-id']: stepId,
    ['approval-id']: approvalId,
  } = await params

  const res = await nueQueryData({
    orgId,
    path: `workflows/${workflowId}/steps/${stepId}/approvals/${approvalId}/contents`,
  })

  return NextResponse.json(res)
}
