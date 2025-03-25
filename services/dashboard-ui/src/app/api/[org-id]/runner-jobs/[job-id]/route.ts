import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getRunnerJob } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (req: NextRequest, { params }: TRouteRes<'org-id' | 'job-id'>) => {
    const orgId = params?.['org-id']
    const runnerJobId = params?.['job-id']

    let job = {}
    try {
      job = await getRunnerJob({ orgId, runnerJobId })
    } catch (error) {
      console.error(error)
    }

    return NextResponse.json(job)
  }
)
