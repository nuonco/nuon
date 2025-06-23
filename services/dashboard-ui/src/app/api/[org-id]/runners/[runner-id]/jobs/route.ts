import { NextRequest, NextResponse } from 'next/server'
import { getRunnerJobs } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'runner-id'>
) => {
  const { searchParams } = new URL(req.url)
  const { ['org-id']: orgId, ['runner-id']: runnerId } = await params

  let jobs = []
  try {
    jobs = await getRunnerJobs({
      orgId,
      runnerId,
      options: Object.fromEntries(searchParams.entries()),
    }).then((res) => res.runnerJobs)
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(jobs)
}
