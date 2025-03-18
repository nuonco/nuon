import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getRunnerLatestHeartbeat } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, runnerId] = req.url.split('/').slice(4, 7)

  let heartbeat = {}
  try {
    heartbeat = await getRunnerLatestHeartbeat({ orgId, runnerId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(heartbeat)
})
