import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getRunnerJob } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, jobId] = req.url.split('/').slice(4, 7)

  let job = {}
  try {
    job = await getRunnerJob({ orgId, jobId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(job)
})
