import { type NextRequest, NextResponse } from 'next/server'
import { getRunnerJobById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'runnerJobId'>
) {
  const { runnerJobId, orgId } = await params

  const response = await getRunnerJobById({
    runnerJobId,
    orgId,
  })
  return NextResponse.json(response)
}
