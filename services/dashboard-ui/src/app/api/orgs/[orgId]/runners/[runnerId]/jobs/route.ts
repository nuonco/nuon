import { type NextRequest, NextResponse } from 'next/server'
import { getRunnerJobs } from '@/lib'
import type { TRouteProps, TRunnerJob } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'runnerId'>
) {
  const { orgId, runnerId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('past-jobs') || undefined
  const groups =
    (searchParams.get('groups') as unknown as TRunnerJob['group'][]) ||
    undefined
  const statuses =
    (searchParams.get('statuses') as unknown as TRunnerJob['status'][]) ||
    undefined

  const response = await getRunnerJobs({
    runnerId,
    orgId,
    limit,
    offset,
    groups,
    statuses,
  })
  return NextResponse.json({
    ...response,
    headers: Object.fromEntries(response.headers?.entries()),
  })
}
