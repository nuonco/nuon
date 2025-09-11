import { type NextRequest, NextResponse } from 'next/server'
import { getWorkflowById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'workflowId'>
) {
  const { orgId, workflowId } = await params

  const response = await getWorkflowById({ orgId, workflowId })
  return NextResponse.json(response)
}
