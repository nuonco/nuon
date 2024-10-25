import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getRunnerLogs } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, runnerId] = req.url.split('/').slice(4, 7)
  const jobId = req.nextUrl.searchParams.get('job_id')

  let logs = {}
  try {
    logs = await getRunnerLogs({ orgId, runnerId, jobId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(logs)
})
