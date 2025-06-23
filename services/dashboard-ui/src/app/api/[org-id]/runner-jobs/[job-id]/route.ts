import { NextRequest, NextResponse } from 'next/server'
import { getRunnerJob } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'job-id'>
) => {
  const { ['org-id']: orgId, ['job-id']: runnerJobId } = await params

  let job = {}
  try {
    job = await getRunnerJob({ orgId, runnerJobId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(job)
}
