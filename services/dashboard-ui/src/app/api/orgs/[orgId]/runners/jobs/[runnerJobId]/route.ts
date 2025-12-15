import { type NextRequest, NextResponse } from 'next/server'
import { getRunnerJob } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'runnerJobId'>
) {
  const { runnerJobId, orgId } = await params

  const response = await getRunnerJob({
    runnerJobId,
    orgId,
  })
  return NextResponse.json(response)
}
