import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getRunnerJobs } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, runnerId] = req.url.split('/').slice(4, 7)

  let jobs = []
  try {
    jobs = await getRunnerJobs({ orgId, runnerId, options: { limit: '1'} }).then(res => res.runnerJobs)
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(jobs)
})
