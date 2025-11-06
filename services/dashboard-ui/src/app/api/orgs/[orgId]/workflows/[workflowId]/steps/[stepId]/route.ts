import { type NextRequest, NextResponse } from 'next/server'
import { getWorkflowStepById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'workflowId' | 'stepId'>
) {
  const { orgId, workflowId, stepId } = await params

  const response = await getWorkflowStepById({
    orgId,
    workflowId,
    workflowStepId: stepId,
  })
  return NextResponse.json(response)
}
