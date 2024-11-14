import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getLogStreamLogs } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, logStreamId] = req.url.split('/').slice(4, 7)
  
  let logs = []
  try {
    logs = await getLogStreamLogs({ orgId, logStreamId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(logs)
})
