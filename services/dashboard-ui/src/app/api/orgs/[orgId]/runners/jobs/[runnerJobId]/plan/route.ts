import { type NextRequest, NextResponse } from 'next/server'
import { getRunnerJobPlanById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | "runnerJobId">
) {
  const { runnerJobId, orgId } = await params

  const response = await getRunnerJobPlanById({
    runnerJobId,
    orgId,
  })  
  return NextResponse.json(response)
}

