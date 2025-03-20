import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getRunnerJobs } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest, { params }) => {
  const { searchParams } = new URL(req.url)
  const orgId = params?.['org-id'] as string
  const runnerId = params?.['runner-id'] as string

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
})
