import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getLogStream } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (
    req: NextRequest,
    { params }: TRouteRes<'org-id' | 'log-stream-id'>
  ) => {
    const orgId = params?.['org-id']
    const logStreamId = params?.['log-stream-id']

    let logStream = {}
    try {
      logStream = await getLogStream({ orgId, logStreamId })
    } catch (error) {
      console.error(error)
    }

    return NextResponse.json(logStream)
  }
)
