import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getRunnerJobs } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (req: NextRequest, { params }: TRouteRes<'org-id' | 'runner-id'>) => {
    const { searchParams } = new URL(req.url)
    const orgId = params?.['org-id']
    const runnerId = params?.['runner-id']

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
)
